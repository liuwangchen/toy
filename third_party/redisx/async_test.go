package redisx

import (
	"context"
	"sync"
	"testing"

	"github.com/liuwangchen/toy/pkg/async"
)

func TestClose(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n, _ := NewAsyncWithConfig(cfg, async.New(func(ctx context.Context, f func()) error {
			f()
			return nil
		}), WithMaxCmdQueue(2048), WithMaxReturnQueue(2048))
		var ready = make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				n.Stop(context.Background())
			}()
		}
		close(ready)
		wg.Wait()
	}
}
