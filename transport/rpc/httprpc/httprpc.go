package httprpc

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/rpc"
)

// Option option
type Option func(s ISetOption)

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) Option {
	return func(o ISetOption) {
		o.SetTimeout(d)
	}
}

// WithConnMiddleware with client middleware.
func WithConnMiddleware(m ...middleware.Middleware) Option {
	return func(o ISetOption) {
		o.SetConnMiddleWare()
	}
}

// WithAddress with server address.
func WithAddress(addr string) Option {
	return func(o ISetOption) {
		o.SetAddr(addr)
	}
}

// WithTLSConfig with tls config.
func WithTLSConfig(c *tls.Config) Option {
	return func(o ISetOption) {
		o.SetTLSConfig(c)
	}
}

type ISetOption interface {
	SetTimeout(timeout time.Duration)
	SetConnMiddleWare(mw ...middleware.Middleware)
	SetAddr(addr string)
	SetTLSConfig(c *tls.Config)
}

func HeaderFromCtx(ctx context.Context) map[string]string {
	result := ServerHeaderFromCtx(ctx)
	if len(result) > 0 {
		return result
	}
	return ClientHeaderFromCtx(ctx)
}

// ServerHeaderFromCtx 获得Header
func ServerHeaderFromCtx(ctx context.Context) map[string]string {
	tr, ok := rpc.FromServerContext(ctx)
	if !ok {
		return nil
	}
	result := make(map[string]string)
	for _, key := range tr.RequestHeader().Keys() {
		result[key] = tr.RequestHeader().Get(key)
	}
	return result
}

// ClientHeaderFromCtx 获得Header
func ClientHeaderFromCtx(ctx context.Context) map[string]string {
	tr, ok := rpc.FromClientContext(ctx)
	if !ok {
		return nil
	}
	result := make(map[string]string)
	for _, key := range tr.RequestHeader().Keys() {
		result[key] = tr.RequestHeader().Get(key)
	}
	return result
}
