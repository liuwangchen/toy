package trace

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/liuwangchen/toy/transport/middleware"
)

type traceLen int

const (
	traceShort traceLen = 0
	traceLong  traceLen = 1
)

// Option is recovery option.
type option func(*options)

type options struct {
	tl traceLen
}

// WithLongLen 设置为长traceId，默认短的
func WithLongLen() option {
	return func(o *options) {
		o.tl = traceLong
	}
}

type traceKey struct {
}

func InjectTraceId(opts ...option) middleware.Middleware {
	op := options{
		tl: traceShort,
	}
	for _, o := range opts {
		o(&op)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			traceId := GetTraceIdFromCtx(ctx)
			if len(traceId) > 0 {
				return handler(ctx, req)
			}
			traceId = strings.ReplaceAll(uuid.New().String(), "-", "")
			switch op.tl {
			case traceShort:
				traceId = traceId[:7]
			case traceLong:
			}
			ctx = ContextWithTraceId(ctx, traceId)
			return handler(ctx, req)
		}
	}
}

func GetTraceIdFromCtx(ctx context.Context) string {
	v := ctx.Value(traceKey{})
	if v == nil {
		return ""
	}
	return v.(string)
}

func ContextWithTraceId(ctx context.Context, traceId string) context.Context {
	if len(traceId) == 0 {
		return ctx
	}
	return context.WithValue(ctx, traceKey{}, traceId)
}

var traceHeaderKey = "traceIdHeaderKey"

func SetTraceIdIntoHeader(header map[string]string, traceId string) {
	if len(traceId) == 0 {
		return
	}
	header[traceHeaderKey] = traceId
}

func GetTraceIdFromHeader(header map[string]string) string {
	return header[traceHeaderKey]
}
