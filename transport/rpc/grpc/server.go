package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/pkg/host"
	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	_ rpc.Endpointer = (*Server)(nil)
)

// ServerOption is gRPC server option.
type ServerOption func(o *Server)

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Middleware with server middleware.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware = m
	}
}

// TLSConfig with TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = c
	}
}

// Listener with server lis
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// UnaryInterceptor returns a ServerOption that sets the UnaryServerInterceptor for the server.
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInts = in
	}
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the server.
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInts = in
	}
}

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// Server is a gRPC server wrapper.
type Server struct {
	*grpc.Server
	baseCtx    context.Context
	tlsConf    *tls.Config
	lis        net.Listener
	network    string
	address    string
	endpoint   *url.URL
	timeout    time.Duration
	middleware []middleware.Middleware
	unaryInts  []grpc.UnaryServerInterceptor
	streamInts []grpc.StreamServerInterceptor
	grpcOpts   []grpc.ServerOption
	health     *health.Server
	ready      bool
}

// NewServer creates a gRPC server by options.
func NewServer(opts ...ServerOption) (*Server, error) {
	srv := &Server{
		baseCtx: context.Background(),
		network: "tcp",
		address: ":0",
		timeout: 1 * time.Second,
		health:  health.NewServer(),
	}
	for _, o := range opts {
		o(srv)
	}
	unaryInts := []grpc.UnaryServerInterceptor{
		srv.unaryServerInterceptor(),
	}
	streamInts := []grpc.StreamServerInterceptor{
		srv.streamServerInterceptor(),
	}
	if len(srv.unaryInts) > 0 {
		unaryInts = append(unaryInts, srv.unaryInts...)
	}
	if len(srv.streamInts) > 0 {
		streamInts = append(streamInts, srv.streamInts...)
	}
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInts...),
		grpc.ChainStreamInterceptor(streamInts...),
	}
	if srv.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
	}
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	// listen and endpoint
	err := srv.listenAndEndpoint()
	if err != nil {
		return nil, err
	}
	// internal register
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server)
	return srv, nil
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	grpc://127.0.0.1:9000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	return s.endpoint, nil
}

// Start start the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	s.baseCtx = ctx
	s.health.Resume()
	logger.Info("[gRPC] server listening on: %s", s.address)
	return s.Serve(s.lis)
}

// Stop stop the gRPC server.
func (s *Server) Stop(ctx context.Context) error {
	s.health.Shutdown()
	s.GracefulStop()
	logger.Info("[gRPC] server stopping")
	return nil
}

func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		_ = s.lis.Close()
		return err
	}
	s.endpoint = endpoint.NewEndpoint("grpc", addr, s.tlsConf != nil)
	s.ready = true
	return nil
}

func (s *Server) Ready() bool {
	return s.ready
}
