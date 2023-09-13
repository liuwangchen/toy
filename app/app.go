package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/pkg/executor"
	"github.com/liuwangchen/toy/pkg/host"
	"github.com/liuwangchen/toy/pkg/ipx"
	"github.com/liuwangchen/toy/registry"
	"github.com/liuwangchen/toy/transport/rpc"
)

// Runner is transport server.
type Runner interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type IReady interface {
	Ready() bool
}

// RunFunc run func
func RunFunc(f func(ctx context.Context) error) Runner {
	return runnerFunc(f)
}

type runnerFunc func(context.Context) error

func (r runnerFunc) Start(ctx context.Context) error {
	return r(ctx)
}

func (r runnerFunc) Stop(ctx context.Context) error {
	return nil
}

// App 应用
type App struct {
	sync.Mutex
	endpoints []*url.URL
	instance  *registry.ServiceInstance

	// options
	id               string
	name             string
	version          string
	metadata         map[string]string
	sigs             []os.Signal
	stopTimeout      time.Duration
	onBeforeF        func() error
	onStopF          func() error
	onRunnerReadyF   func() error
	runners          []Runner
	registrar        registry.Registrar
	registrarTimeout time.Duration
	pprofAddr        string
}

// Option option
type Option func(o *App)

// WithID ID with service id.
func WithID(id string) Option {
	return func(o *App) { o.id = id }
}

// WithName Name with service name.
func WithName(name string) Option {
	return func(o *App) { o.name = name }
}

// WithVersion Version with service version.
func WithVersion(version string) Option {
	return func(o *App) { o.version = version }
}

// WithMetadata Metadata with service metadata.
func WithMetadata(md map[string]string) Option {
	return func(o *App) { o.metadata = md }
}

// WithSignal Signal 信号
func WithSignal(sigs ...os.Signal) Option {
	return func(o *App) { o.sigs = sigs }
}

// WithStopTimeout StopTimeout 关闭超时
func WithStopTimeout(t time.Duration) Option {
	return func(o *App) { o.stopTimeout = t }
}

func WithOnBeforeFunc(f func() error) Option {
	return func(o *App) { o.onBeforeF = f }
}

func WithOnStopFunc(f func() error) Option {
	return func(o *App) { o.onStopF = f }
}

func WithOnRunnerReadyFunc(f func() error) Option {
	return func(o *App) { o.onRunnerReadyF = f }
}

// WithRunners Runners 服务
func WithRunners(s ...Runner) Option {
	return func(o *App) {
		o.runners = append(o.runners, s...)
	}
}

// WithRegistrar Registrar with service registry.
func WithRegistrar(r registry.Registrar) Option {
	return func(o *App) { o.registrar = r }
}

// WithRegistrarTimeout RegistrarTimeout with registrar timeout.
func WithRegistrarTimeout(t time.Duration) Option {
	return func(o *App) { o.registrarTimeout = t }
}

func WithPProf(addr string) Option {
	return func(o *App) { o.pprofAddr = addr }
}

// New 构造
func New(opts ...Option) *App {
	a := &App{
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: 10 * time.Second,
		stopTimeout:      time.Second * 30,
	}
	if id, err := uuid.NewUUID(); err == nil {
		a.id = id.String()
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// ID returns app instance id.
func (a *App) ID() string { return a.id }

// Name returns service name.
func (a *App) Name() string { return a.name }

// Version returns app version.
func (a *App) Version() string { return a.version }

// Metadata returns service metadata.
func (a *App) Metadata() map[string]string { return a.metadata }

// Endpoint returns endpoints.
func (a *App) Endpoint() []string {
	if a.instance == nil {
		return []string{}
	}
	return a.instance.Endpoints
}

func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)
	for _, srv := range a.runners {
		if r, ok := srv.(rpc.Endpointer); ok {
			e, err := r.Endpoint()
			if err != nil {
				return nil, err
			}
			endPointStr, _ := url.QueryUnescape(e.String())
			endpoints = append(endpoints, endPointStr)
		}
	}
	return &registry.ServiceInstance{
		ID:         a.ID(),
		Name:       a.Name(),
		Version:    a.Version(),
		Metadata:   a.Metadata(),
		Endpoints:  endpoints,
		LaunchTime: time.Now().Unix(),
		Ip:         ipx.GetOutboundIP(),
	}, nil
}

func (a *App) before(ctx context.Context) error {
	if a.onBeforeF != nil {
		err := a.onBeforeF()
		if err != nil {
			return err
		}
	}
	if a.registrar == nil {
		return nil
	}
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}
	rctx, rcancel := context.WithTimeout(ctx, a.registrarTimeout)
	defer rcancel()
	if err := a.registrar.Register(rctx, instance); err != nil {
		return err
	}
	a.instance = instance
	return nil
}

func (a *App) onStop(ctx context.Context) error {
	if a.onStopF != nil {
		err := a.onStopF()
		if err != nil {
			return err
		}
	}
	if a.registrar == nil {
		return nil
	}
	if a.instance == nil {
		return nil
	}
	a.Lock()
	defer a.Unlock()
	ctx1, cancel := context.WithTimeout(ctx, a.registrarTimeout)
	defer cancel()
	if err := a.registrar.Deregister(ctx1, a.instance); err != nil {
		return err
	}
	return nil
}

// 当所有runner都ready后，执行回调
func (a *App) checkRunnerReady(_ context.Context) error {
	if a.onRunnerReadyF == nil {
		return nil
	}
	isAllRunnerReadyF := func() bool {
		for _, runner := range a.runners {
			ready, ok := runner.(IReady)
			if !ok {
				continue
			}
			if !ready.Ready() {
				return false
			}
		}
		return true
	}
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()
	for range ticker.C {
		if isAllRunnerReadyF() {
			return a.onRunnerReadyF()
		}
	}
	return nil
}

type pprofRunner struct {
	Addr     string
	server   *http.Server
	endPoint *url.URL
	lis      net.Listener
}

func (this *pprofRunner) Endpoint() (*url.URL, error) {
	return this.endPoint, nil
}

func newPprofRunner(addr string) Runner {
	lis, _ := net.Listen("tcp", addr)
	addr, _ = host.Extract(addr, lis)
	return &pprofRunner{Addr: addr, lis: lis, endPoint: endpoint.NewEndpoint("pprof", addr, false)}
}

func (this *pprofRunner) Start(ctx context.Context) error {
	if this.lis == nil {
		return errors.New("lis is nil")
	}
	server := &http.Server{
		Addr:    this.Addr,
		Handler: http.DefaultServeMux,
	}

	this.server = server
	err := server.Serve(this.lis)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (this *pprofRunner) Stop(ctx context.Context) error {
	profileCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := this.server.Shutdown(profileCtx)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Run(rootCtx context.Context) error {
	// pprof
	if len(a.pprofAddr) > 0 {
		a.runners = append(a.runners, newPprofRunner(a.pprofAddr))
	}
	ctx, cancel := context.WithCancel(rootCtx)
	// 开始
	runStarts := make([]executor.Executor, 0, len(a.runners))
	for _, runner := range a.runners {
		runStarts = append(runStarts, executor.Func(runner.Start))
	}
	// 结束
	runStops := make([]executor.Executor, 0, len(a.runners))
	for _, runner := range a.runners {
		runStops = append(runStops, executor.Func(runner.Stop))
	}
	return executor.Execute(ctx,
		// 并行开始
		executor.Parallel(
			executor.Parallel(runStarts...),
			executor.Func(a.checkRunnerReady),
		),
		// 开始前执行函数
		executor.WithBefore(executor.Func(a.before)),
		executor.WithAfter(executor.Append(
			// 并行关闭
			executor.Parallel(runStops...),
			// app的最后执行函数
			executor.Func(a.onStop),
		)),
		// 信号量
		executor.WithSignal(executor.Func(func(ctx context.Context) error {
			cancel()
			return nil
		}), a.sigs...))
}
