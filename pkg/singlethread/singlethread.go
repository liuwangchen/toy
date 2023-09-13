package singlethread

import (
	"context"
	"time"

	"github.com/liuwangchen/toy/pkg/async"
)

const (
	defaultFChanBuf = 100
)

// ST 单协程驱动single thread
type ST struct {
	fChan    chan func()         // 在一个协程里执行的func
	tickF    func(now time.Time) // 逻辑tick
	ticker   *time.Ticker        // 逻辑ticker
	fChanBuf int
	as       async.IAsync
}

type Option func(st *ST)

func WithTickF(tickF func(now time.Time)) Option {
	return func(s *ST) {
		s.tickF = tickF
	}
}

func WithFChanBuf(bufN int) Option {
	return func(s *ST) {
		s.fChanBuf = bufN
	}
}

func NewST(opts ...Option) *ST {
	st := &ST{
		fChanBuf: defaultFChanBuf,
	}
	for _, opt := range opts {
		opt(st)
	}
	if st.tickF != nil {
		st.ticker = time.NewTicker(time.Millisecond * 100)
	}
	st.fChan = make(chan func(), st.fChanBuf)
	st.as = async.New(func(ctx context.Context, f func()) error {
		st.fChan <- f
		return nil
	})
	return st
}

func (this *ST) getTickerC() <-chan time.Time {
	if this.ticker == nil {
		return nil
	}
	return this.ticker.C
}

// Async 返回fchan路的async
func (this *ST) Async() async.IAsync {
	return this.as
}

func (this *ST) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done(): // ctx
			return nil
		case now := <-this.getTickerC(): // logicTicker
			this.tickF(now)
		case f := <-this.fChan: // async
			f()
		}
	}
}

func (this *ST) Stop(ctx context.Context) error {
	if this.ticker != nil {
		this.ticker.Stop()
	}
	return nil
}
