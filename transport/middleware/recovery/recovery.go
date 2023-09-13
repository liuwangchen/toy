package recovery

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/transport/middleware"
)

// ErrUnknownRequest is unknown request error.
var ErrUnknownRequest = errors.New("unknown request error")

// HandlerFunc is recovery handler func.
type HandlerFunc func(ctx context.Context, req, err interface{}) error

// Option is recovery option.
type Option func(*options)

type options struct {
	handler HandlerFunc
}

// WithHandler with recovery handler.
func WithHandler(h HandlerFunc) Option {
	return func(o *options) {
		o.handler = h
	}
}

// Recovery is a server middleware that recovers from any panics.
func Recovery(opts ...Option) middleware.Middleware {
	op := options{
		handler: func(ctx context.Context, req, err interface{}) error {
			trace := string(debug.Stack())
			logger.Fatal("Recovery panic=%v stack=%s", err, trace)
			return ErrUnknownRequest
		},
	}
	for _, o := range opts {
		o(&op)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if rerr := recover(); rerr != nil {
					err = op.handler(ctx, req, rerr)
				}
			}()
			return handler(ctx, req)
		}
	}
}
