package natsrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/liuwangchen/toy/transport/middleware"
)

const (
	PROTOBUF_ENCODER = "protobuf"
	JSON_ENCODER     = "json"
	GOB_ENCODER      = "gob"
	DEFAULT_ENCODER  = "default"
)

// Option option
type Option func(s ISetOption)

// WithConn 连接
func WithConn(conn *nats.Conn) Option {
	return func(s ISetOption) {
		s.SetConn(conn)
	}
}

// WithEncType 序列化类型
func WithEncType(encType string) Option {
	return func(s ISetOption) {
		s.SetEncodeType(encType)
	}
}

// WithNatsOption nats相关option
func WithNatsOption(natsOptions ...nats.Option) Option {
	return func(s ISetOption) {
		s.SetNatsOption(natsOptions...)
	}
}

// WithConnMiddleware 中间件
func WithConnMiddleware(mw ...middleware.Middleware) Option {
	return func(s ISetOption) {
		s.SetConnMiddleWare(mw...)
	}
}

// WithNamespace 命名空间
func WithNamespace(ns string) Option {
	return func(s ISetOption) {
		s.SetNamespace(ns)
	}
}

// WithTimeout 默认call超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(s ISetOption) {
		s.SetTimeout(timeout)
	}
}

// WithAddr 默认call超时时间
func WithAddr(addr string) Option {
	return func(s ISetOption) {
		s.SetAddr(addr)
	}
}

type ISetOption interface {
	SetConn(conn *nats.Conn)
	SetEncodeType(encType string)
	SetNatsOption(natsOptions ...nats.Option)
	SetConnMiddleWare(mw ...middleware.Middleware)
	SetNamespace(namespace string)
	SetTimeout(timeout time.Duration)
	SetAddr(addr string)
}

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

type headerKey struct{}

// WithHeaderContext 填充Header
func WithHeaderContext(ctx context.Context, addHeader map[string]string) context.Context {
	header := HeaderFromCtx(ctx)
	if len(header) == 0 {
		header = map[string]string{}
	}
	for key, v := range addHeader {
		header[key] = v
	}
	return context.WithValue(ctx, headerKey{}, header)
}

// HeaderFromCtx 获得Header
func HeaderFromCtx(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(headerKey{})
	if val == nil {
		return nil
	}
	return val.(map[string]string)
}

type callTopicKey struct{}

// WithCallTopicContext 填充Header
func WithCallTopicContext(ctx context.Context, callTopic interface{}) context.Context {
	newCtx := context.WithValue(ctx, callTopicKey{}, callTopic)
	return newCtx
}

// CallTopicFromCtx 获得Header
func CallTopicFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(callTopicKey{})
	if val == nil {
		return ""
	}
	return fmt.Sprint(val)
}
