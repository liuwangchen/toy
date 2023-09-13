package rpc

import (
	"context"
	"net/url"

	// init encoding
	_ "github.com/liuwangchen/toy/transport/encoding/form"
	_ "github.com/liuwangchen/toy/transport/encoding/json"
	_ "github.com/liuwangchen/toy/transport/encoding/proto"
	_ "github.com/liuwangchen/toy/transport/encoding/xml"
	_ "github.com/liuwangchen/toy/transport/encoding/yaml"
)

// Endpointer is registry endpoint.
type Endpointer interface {
	Endpoint() (*url.URL, error)
}

// Header is the storage medium used by a Header.
type Header interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

// Transporter is transport context value interface.
type Transporter interface {
	// Kind transporter
	// grpc
	// http
	Kind() Kind
	// Endpoint return server or client endpoint
	// Server Transport: grpc://127.0.0.1:9000
	// Client Transport: discovery:///provider-demo
	Endpoint() string
	// Operation Service full method selector generated by protobuf
	// example: /helloworld.Greeter/SayHello
	Operation() string
	// RequestHeader return transport request header
	// http: http.Header
	// grpc: metadata.MD
	RequestHeader() Header
	// ReplyHeader return transport reply/response header
	// only valid for server transport
	// http: http.Header
	// grpc: metadata.MD
	ReplyHeader() Header
}

// Kind defines the type of Transport
type Kind string

func (k Kind) String() string { return string(k) }

// Defines a set of transport kind
const (
	KindGRPC  Kind = "grpc"
	KindHTTP  Kind = "http"
	KindNats  Kind = "nats"
	KindUdp   Kind = "udp"
	KindKafka Kind = "kafka"
)

type (
	serverTransportKey struct{}
	clientTransportKey struct{}
)

// NewServerContext returns a new Context that carries value.
func NewServerContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, serverTransportKey{}, tr)
}

// FromServerContext returns the Transport value stored in ctx, if any.
func FromServerContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(serverTransportKey{}).(Transporter)
	return
}

// NewClientContext returns a new Context that carries value.
func NewClientContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, clientTransportKey{}, tr)
}

// FromClientContext returns the Transport value stored in ctx, if any.
func FromClientContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(clientTransportKey{}).(Transporter)
	return
}
