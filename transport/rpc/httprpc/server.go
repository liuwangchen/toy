package httprpc

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/pkg/host"
	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/rpc"
)

var (
	_ rpc.Endpointer = (*ServerConn)(nil)
)

// WithServerNetwork with server network.
func WithServerNetwork(network string) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.network = network
	}
}

// WithServerFilter with HTTP middleware option.
func WithServerFilter(filters ...FilterFunc) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.filters = filters
	}
}

// WithServerRequestDecoder with request decoder.
func WithServerRequestDecoder(dec DecodeRequestFunc) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.dec = dec
	}
}

// WithServerResponseEncoder with response encoder.
func WithServerResponseEncoder(en EncodeResponseFunc) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.enc = en
	}
}

// WithServerErrorEncoder with error encoder.
func WithServerErrorEncoder(en EncodeErrorFunc) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.ene = en
	}
}

// WithServerStrictSlash is with mux's WithServerStrictSlash
// If true, when the path pattern is "/path/", accessing "/path" will
// redirect to the former and vice versa.
func WithServerStrictSlash(strictSlash bool) Option {
	return func(o ISetOption) {
		server, ok := o.(*ServerConn)
		if !ok {
			return
		}
		server.strictSlash = strictSlash
	}
}

// WithServerListener with server lis
func WithServerListener(lis net.Listener) Option {
	return func(s ISetOption) {
		server, ok := s.(*ServerConn)
		if !ok {
			return
		}
		server.lis = lis
	}
}

// ServerConn is an HTTP server wrapper.
type ServerConn struct {
	*http.Server
	lis         net.Listener
	tlsConf     *tls.Config
	endpoint    *url.URL
	network     string
	address     string
	timeout     time.Duration
	filters     []FilterFunc
	ms          []middleware.Middleware
	dec         DecodeRequestFunc
	enc         EncodeResponseFunc
	ene         EncodeErrorFunc
	strictSlash bool
	router      *mux.Router
	ready       bool
}

func (s *ServerConn) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

func (s *ServerConn) SetConnMiddleWare(mw ...middleware.Middleware) {
	s.ms = mw
}

func (s *ServerConn) SetAddr(addr string) {
	s.address = addr
}

func (s *ServerConn) SetTLSConfig(c *tls.Config) {
	s.TLSConfig = c
}

// NewServerConn creates an HTTP server by options.
func NewServerConn(opts ...Option) (*ServerConn, error) {
	srv := &ServerConn{
		network:     "tcp",
		address:     ":80",
		timeout:     1 * time.Second,
		dec:         DefaultRequestDecoder,
		enc:         DefaultResponseEncoder,
		ene:         DefaultErrorEncoder,
		strictSlash: true,
	}
	for _, o := range opts {
		o(srv)
	}
	srv.router = mux.NewRouter().StrictSlash(srv.strictSlash)
	srv.router.Use(srv.filter())
	srv.Server = &http.Server{
		Handler:   FilterChain(srv.filters...)(srv.router),
		TLSConfig: srv.tlsConf,
	}
	err := srv.listenAndEndpoint()
	if err != nil {
		return nil, err
	}
	return srv, nil
}

// Route registers an HTTP router.
func (s *ServerConn) Route(prefix string, filters ...FilterFunc) *Router {
	return newRouter(prefix, s, filters...)
}

// Handle registers a new route with a matcher for the URL path.
func (s *ServerConn) Handle(path string, h http.Handler) {
	s.router.Handle(path, h)
}

// HandlePrefix registers a new route with a matcher for the URL path prefix.
func (s *ServerConn) HandlePrefix(prefix string, h http.Handler) {
	s.router.PathPrefix(prefix).Handler(h)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *ServerConn) HandleFunc(path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h)
}

// HandleHeader registers a new route with a matcher for the header.
func (s *ServerConn) HandleHeader(key, val string, h http.HandlerFunc) {
	s.router.Headers(key, val).Handler(h)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *ServerConn) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(res, req)
}

func (s *ServerConn) filter() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
			} else {
				ctx, cancel = context.WithCancel(req.Context())
			}
			defer cancel()

			pathTemplate := req.URL.Path
			if route := mux.CurrentRoute(req); route != nil {
				// /path/123 -> /path/{id}
				pathTemplate, _ = route.GetPathTemplate()
			}

			tr := &Transport{
				endpoint:     s.endpoint.String(),
				operation:    pathTemplate,
				reqHeader:    headerCarrier(req.Header),
				replyHeader:  headerCarrier(w.Header()),
				request:      req,
				pathTemplate: pathTemplate,
			}

			tr.request = req.WithContext(rpc.NewServerContext(ctx, tr))
			next.ServeHTTP(w, tr.request)
		})
	}
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	http://127.0.0.1:8000?isSecure=false
func (s *ServerConn) Endpoint() (*url.URL, error) {
	return s.endpoint, nil
}

// Start start the HTTP server.
func (s *ServerConn) Start(ctx context.Context) error {
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	logger.Info("[HTTP] server listening on: %s", s.address)
	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.lis, "", "")
	} else {
		err = s.Serve(s.lis)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop stop the HTTP server.
func (s *ServerConn) Stop(ctx context.Context) error {
	logger.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}

func (s *ServerConn) listenAndEndpoint() error {
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
	s.endpoint = endpoint.NewEndpoint("http", addr, s.tlsConf != nil)
	s.ready = true
	return nil
}

type ServiceOpt struct {
	Mw []middleware.Middleware
}

// ServiceOption service option
type ServiceOption func(options *ServiceOpt)

// WithServiceMiddleware 超时时间
func WithServiceMiddleware(mw ...middleware.Middleware) ServiceOption {
	return func(s *ServiceOpt) {
		s.Mw = mw
	}
}

func (s *ServerConn) Ready() bool {
	return s.ready
}
