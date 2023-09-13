package redisx

import (
	"context"
	"fmt"
	"hash/crc32"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/liuwangchen/toy/pkg/async"
)

// Cmder redis 命令(无需回调)
type Cmder interface {
	Do(cli *Client) error     // 执行命令
	Key() string              // key
	Context() context.Context // context
}

// CallbackCmder 带回调的redis命令接口
type CallbackCmder interface {
	Cmder
	Callback()
}

type (
	hashFun      func(string, int) int
	ErrorHandler func(error) // 错误处理
)

// 选项
type option struct {
	concurrency    int
	hash           hashFun
	maxCmdQueue    int // 每个thread最大命令队列
	maxReturnQueue int // 最大返回队列
	errorHandler   ErrorHandler
	onStop         []func()
	overtimeDur    time.Duration // 超时时间
}
type OpOption func(*option)

func defaultOption() option {
	return option{
		maxCmdQueue:    1024,
		maxReturnQueue: 1024,
		concurrency:    8,
		overtimeDur:    100 * time.Millisecond,
		hash: func(key string, count int) int {
			if count == 1 {
				return 0
			}
			return int(crc32.ChecksumIEEE([]byte(key)) % uint32(count))
		},
		errorHandler: func(err error) {
			if _, ok := err.(net.Error); ok {
				fmt.Printf("[AsyncClient] net err[%v]\n", err)
			}
		},
	}
}

func (o *option) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithMaxCmdQueue 设置最大命令队列
func WithMaxCmdQueue(size int) OpOption {
	return func(opt *option) {
		opt.maxCmdQueue = size
	}
}

// WithMaxReturnQueue 设置最大返回队列
func WithMaxReturnQueue(size int) OpOption {
	return func(opt *option) {
		opt.maxReturnQueue = size
	}
}

func WithStopHandler(f func()) OpOption {
	return func(opt *option) {
		opt.onStop = append(opt.onStop, f)
	}
}

// WithOverTimeDurWarn 超时报警时间
func WithOverTimeDurWarn(t time.Duration) OpOption {
	return func(opt *option) {
		opt.overtimeDur = t
	}
}

// WithConcurrency 设置最大并发数
func WithConcurrency(n int) OpOption {
	return func(opt *option) {
		opt.concurrency = n
	}
}

// WithHashFunc 设置最大返回队列
func WithHashFunc(hash hashFun) OpOption {
	return func(opt *option) {
		opt.hash = hash
	}
}

type cmdClient struct {
	cmds chan Cmder
}

// AsyncClient redis 异步调用
type AsyncClient struct {
	cli      *Client
	worker   []*cmdClient
	out      chan CallbackCmder
	stopChan chan struct{}
	opt      option
	wg       sync.WaitGroup
	once     sync.Once
	as       async.IAsync
}

func NewAsyncWithConfig(cfg Config, as async.IAsync, opts ...OpOption) (*AsyncClient, error) {
	cli, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithStopHandler(func() {
		if err := cli.Close(); err != nil {
			fmt.Printf("[NewAsyncWithConfig] err [%v]", err)
		}
	}))
	return NewAsync(cli, as, opts...), nil
}

func NewAsync(cli *Client, as async.IAsync, opts ...OpOption) *AsyncClient {
	var o = defaultOption()
	o.applyOpts(opts)
	c := &AsyncClient{
		out:      make(chan CallbackCmder, o.maxReturnQueue),
		stopChan: make(chan struct{}),
		opt:      o,
		cli:      cli,
		as:       as,
	}

	for i := 0; i < o.concurrency; i++ {
		worker := c.newCmdClient(o.maxCmdQueue)
		c.worker = append(c.worker, worker)
	}

	// 启动所有cmdClient
	c.wg.Add(1)
	go c.runClient()

	c.wg.Add(1)
	go c.pushCallback()
	return c
}

func (c *AsyncClient) runClient() {
	defer c.wg.Done()
	defer close(c.out)

	var cliWg sync.WaitGroup
	for _, w := range c.worker {
		cliWg.Add(1)
		go func(cw *cmdClient) {
			defer cliWg.Done()
			c.serve(cw)
		}(w)
	}

	cliWg.Wait()
}

func (c *AsyncClient) newCmdClient(maxCmdQueue int) *cmdClient {
	return &cmdClient{
		cmds: make(chan Cmder, maxCmdQueue),
	}
}

// Sync 同步调用
func (c *AsyncClient) Sync() *Client {
	return c.cli
}

func (c *AsyncClient) Out() <-chan CallbackCmder {
	return c.out
}

func (c *AsyncClient) pushCallback() {
	defer c.wg.Done()
	for cb := range c.Out() {
		c.as.DoWithNoRet(cb.Context(), cb.Callback)
	}
}

func (c *AsyncClient) serve(w *cmdClient) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("panic: %v: %v", err, buf)
		}
	}()

LOOP:
	for {
		select {
		case cmd := <-w.cmds:
			begin := time.Now()
			if err := cmd.Do(c.cli); err != nil {
				fmt.Printf("AsyncClient: server cmd[%v] %s", cmd.Key(), err)
				if c.opt.errorHandler != nil {
					c.opt.errorHandler(err)
				}
			} else {
				executeTime := time.Since(begin)
				if executeTime > c.opt.overtimeDur {
					fmt.Printf("[AsyncClient] cmd[%v] execute slow, time[%vms]", cmd.Key(), executeTime.Milliseconds())
				}
			}
			if cbCmd, ok := cmd.(CallbackCmder); ok {
				c.out <- cbCmd
			}
		case <-c.stopChan:
			break LOOP
		}
	}

	for {
		select {
		case cmd := <-w.cmds:
			if err := cmd.Do(c.cli); err != nil {
				fmt.Printf("redisasync: exit server cmd[%v] %s", cmd.Key(), err)
			}
		default:
			return
		}
	}
}

// Stop 关闭
func (c *AsyncClient) Stop(ctx context.Context) (err error) {
	c.once.Do(func() {
		close(c.stopChan)
		over := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(over)
		}()
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-over:
			break
		}
	})
	return
}

func (c *AsyncClient) addCmd(cmd Cmder) {
	index := c.opt.hash(cmd.Key(), len(c.worker))
	c.worker[index].cmds <- cmd
}

// DoCustomCmd 执行自定义命令
func (c *AsyncClient) DoCustomCmd(redisCmd Cmder) {
	c.addCmd(redisCmd)
}

func (c *AsyncClient) Set(ctx context.Context, key string, value interface{}) error {
	req, err := Encode(value)
	if err != nil {
		return err
	}
	c.addCmd(&setCmd{
		ctx:   ctx,
		key:   key,
		value: req,
	})
	return nil
}

func (c *AsyncClient) Do(ctx context.Context, cb func(*StringRet), args ...interface{}) error {
	var encodeArgs = make([]interface{}, 0, len(args))
	for _, arg := range args {
		v, err := Encode(arg)
		if err != nil {
			return err
		}
		encodeArgs = append(encodeArgs, v)
	}
	c.addCmd(&doCmd{
		ctx:      ctx,
		cmd:      encodeArgs,
		callback: cb,
	})
	return nil
}

func (c *AsyncClient) Get(ctx context.Context, key string, cb func(*StringRet)) {
	c.addCmd(&getCmd{
		ctx:      ctx,
		key:      key,
		callback: cb,
	})
}

func (c *AsyncClient) Del(ctx context.Context, key string) {
	c.addCmd(&delCmd{
		ctx: ctx,
		key: key,
	})
}

func (c *AsyncClient) HMSet(ctx context.Context, key string, value MapReq) {
	c.addCmd(&hmsetCmd{
		ctx:   ctx,
		key:   key,
		value: value,
	})
}

func (c *AsyncClient) HMGet(ctx context.Context, key string, field []string, cb func(ret *StringSliceRet)) {
	c.addCmd(&hmgetCmd{
		ctx:      ctx,
		key:      key,
		field:    field,
		callback: cb,
	})
}

func (c *AsyncClient) HDel(ctx context.Context, key string, field ...string) {
	c.addCmd(&hdelCmd{
		ctx:   ctx,
		key:   key,
		field: field,
	})
}

func (c *AsyncClient) SetNX(ctx context.Context, key string, value interface{}, cb func(ok bool)) error {
	req, err := Encode(value)
	if err != nil {
		return err
	}
	c.addCmd(&setNXCmd{
		ctx:      ctx,
		key:      key,
		value:    req,
		callback: cb,
	})
	return nil
}

func (c *AsyncClient) SetNXAndReturn(ctx context.Context, key string, value interface{}, cb func(ok bool, v string)) error {
	req, err := Encode(value)
	if err != nil {
		return err
	}
	c.addCmd(&setNXReturnCmd{
		ctx:      ctx,
		key:      key,
		value:    req,
		callback: cb,
	})
	return nil
}

func (c *AsyncClient) SetNXEXAndReturn(ctx context.Context, key string, value interface{}, expire time.Duration, cb func(ok bool, v string)) error {
	req, err := Encode(value)
	if err != nil {
		return err
	}
	c.addCmd(&setNXEXReturnCmd{
		ctx:      ctx,
		key:      key,
		value:    req,
		expire:   expire,
		callback: cb,
	})
	return nil
}
