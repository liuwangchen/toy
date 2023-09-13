package async

import (
	"context"
	"runtime/debug"

	"github.com/liuwangchen/toy/logger"
)

// IAsync 异步执行
type IAsync interface {
	// Do 异步执行 在其他协程调主协程函数,阻塞等待f的返回值
	// f 必须是主协程函数
	// 不能在fc的消费协程调，只能在其他协程调
	Do(ctx context.Context, f func() (interface{}, error)) (ret interface{}, err error)

	// DoWithNoRet 异步执行 在其他协程调主协程函数,阻塞等待f,不需要返回值
	// f 必须是主协程函数
	// 不能在fc的消费协程调，只能在其他协程调
	DoWithNoRet(ctx context.Context, f func())

	// AsyncDo 异步执行 在其他协程调主协程函数,阻塞等待f的返回值
	// f 必须是主协程函数，cb是结果回调
	// 不能在fc的消费协程调，只能在其他协程调
	AsyncDo(ctx context.Context, f func(cb func(interface{}, error))) (interface{}, error)

	// Go 执行 主协程起协程执行完并回调
	// g 是新起协程执行的函数
	// cb 是回调到主协程函数
	// 本函数必须在主协程调用
	Go(ctx context.Context, g func(context.Context) (interface{}, error), cb func(interface{}, error))
}

// PushFunc 获得fc
type PushFunc func(context.Context, func()) error

type Middleware func(context.Context, func()) func()

type option struct {
	panicHandler func(interface{})
	fms          []Middleware
}

type Option func(*option)

// WithPanicHandler panic handler
func WithPanicHandler(fn func(interface{})) Option {
	return func(o *option) {
		o.panicHandler = fn
	}
}

func WithMiddleware(fms ...Middleware) Option {
	return func(o *option) {
		o.fms = fms
	}
}

// chain returns a Middleware that specifies the chained handler for endpoint.
func chain(m ...Middleware) Middleware {
	return func(ctx context.Context, next func()) func() {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](ctx, next)
		}
		return next
	}
}

// Async 异步执行
type Async struct {
	opt option
	pf  PushFunc
}

// New 构造
func New(pf PushFunc, opts ...Option) *Async {
	opt := defaultOpt()
	for _, v := range opts {
		v(&opt)
	}
	return &Async{
		opt: opt,
		pf:  pf,
	}
}

var (
	// Default 默认async实现
	Default = New(func(ctx context.Context, f func()) error {
		f()
		return nil
	})
)

// CustomAsync 通过context来自定义async
func CustomAsync(cf func(context.Context) (IAsync, error), opts ...Option) IAsync {
	return New(func(ctx context.Context, f func()) error {
		as, err := cf(ctx)
		if err != nil {
			return err
		}
		as.DoWithNoRet(ctx, f)
		return nil
	}, opts...)
}

// Do 异步执行 在其他协程调主协程函数,阻塞等待f的返回值
// f 必须是主协程函数
// 不能在fc的消费协程调，只能在其他协程调
func (a *Async) Do(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	done := make(chan struct{})
	var (
		ret interface{}
		err error
	)
	f1 := func() {
		if a.opt.panicHandler != nil {
			defer func() {
				if e := recover(); e != nil {
					a.opt.panicHandler(e)
				}
			}()
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
			ret, err = f()
			close(done)
		}
	}

	// 中间件
	if len(a.opt.fms) > 0 {
		f1 = chain(a.opt.fms...)(ctx, f1)
	}

	if err1 := a.pf(ctx, f1); err1 != nil {
		return nil, err1
	}
	select {
	case <-done:
		return ret, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// DoWithNoRet 异步执行 在其他协程调主协程函数,阻塞等待f,不需要返回值
// f 必须是主协程函数
// 不能在fc的消费协程调，只能在其他协程调
func (a *Async) DoWithNoRet(ctx context.Context, f func()) {
	f1 := func() (interface{}, error) {
		f()
		return nil, nil
	}
	_, _ = a.Do(ctx, f1)
}

// AsyncDo 异步执行 在其他协程调主协程函数,阻塞等待f的返回值
// f 必须是主协程函数，cb是结果回调
// 不能在fc的消费协程调，只能在其他协程调
func (a *Async) AsyncDo(ctx context.Context, f func(cb func(interface{}, error))) (interface{}, error) {
	done := make(chan struct{})
	var (
		ret interface{}
		err error
	)
	cb := func(_ret interface{}, _err error) {
		ret, err = _ret, _err
		close(done)
	}
	a.DoWithNoRet(ctx, func() {
		f(cb)
	})
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return nil, err
	case <-done:
		return ret, err
	}
}

// Go 执行 主协程起协程执行完并回调
// g 是新起协程执行的函数
// cb 是回调到主协程函数
// 本函数必须在主协程调用
func (a *Async) Go(ctx context.Context, g func(context.Context) (interface{}, error), cb func(interface{}, error)) {
	go func() {
		ret, err := g(ctx)
		a.DoWithNoRet(ctx, func() {
			cb(ret, err)
		})
	}()
}

func defaultOpt() option {
	return option{
		panicHandler: func(e interface{}) {
			trace := debug.Stack()
			logger.Fatal("[Async] panic[%v]\nstack:%s", e, string(trace))
		},
	}
}
