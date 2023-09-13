package httprpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/liuwangchen/toy/pkg/copier"
	"github.com/liuwangchen/toy/pkg/httputil"
	"github.com/liuwangchen/toy/registry"
	"github.com/liuwangchen/toy/selector"
	"github.com/liuwangchen/toy/selector/wrr"
	"github.com/liuwangchen/toy/transport/encoding"
	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/rpc"
)

// DecodeErrorFunc is decode error func.
type DecodeErrorFunc func(ctx context.Context, res *http.Response) error

// EncodeRequestFunc is request encode func.
type EncodeRequestFunc func(ctx context.Context, contentType string, in interface{}) (body []byte, err error)

// DecodeResponseFunc is response decode func.
type DecodeResponseFunc func(ctx context.Context, res *http.Response, out interface{}) error

// WithClientTransport with client transport.
func WithClientTransport(trans http.RoundTripper) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.transport = trans
	}
}

// WithClientUserAgent with client user agent.
func WithClientUserAgent(ua string) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.userAgent = ua
	}
}

// WithClientRequestEncoder with client request encoder.
func WithClientRequestEncoder(encoder EncodeRequestFunc) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.encoder = encoder
	}
}

// WithClientResponseDecoder with client response decoder.
func WithClientResponseDecoder(decoder DecodeResponseFunc) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.decoder = decoder
	}
}

// WithClientErrorDecoder with client error decoder.
func WithClientErrorDecoder(errorDecoder DecodeErrorFunc) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.errorDecoder = errorDecoder
	}
}

// WithClientDiscovery with client discovery.
func WithClientDiscovery(d registry.Discovery) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.discovery = d
	}
}

// WithClientSelector with client selector.
func WithClientSelector(selector selector.Selector) Option {
	return func(o ISetOption) {
		c, ok := o.(*ClientConn)
		if !ok {
			return
		}
		c.selector = selector
	}
}

// ClientConn is an HTTP client.
type ClientConn struct {
	ctx          context.Context
	tlsConf      *tls.Config
	timeout      time.Duration
	endpoint     string
	userAgent    string
	encoder      EncodeRequestFunc
	decoder      DecodeResponseFunc
	errorDecoder DecodeErrorFunc
	transport    http.RoundTripper
	selector     selector.Selector
	discovery    registry.Discovery
	middleware   []middleware.Middleware
	target       *Target
	r            *resolver
	cc           *http.Client
	streamCc     *http.Client
	insecure     bool
}

func (conn *ClientConn) SetTimeout(timeout time.Duration) {
	conn.timeout = timeout
}

func (conn *ClientConn) SetConnMiddleWare(mw ...middleware.Middleware) {
	conn.middleware = mw
}

func (conn *ClientConn) SetAddr(addr string) {
	conn.endpoint = addr
}

func (conn *ClientConn) SetTLSConfig(c *tls.Config) {
	conn.tlsConf = c
}

// NewClientConn returns an HTTP client.
func NewClientConn(ctx context.Context, opts ...Option) (*ClientConn, error) {
	conn := &ClientConn{
		ctx:          ctx,
		timeout:      2000 * time.Millisecond,
		encoder:      DefaultRequestEncoder,
		decoder:      DefaultResponseDecoder,
		errorDecoder: DefaultErrorDecoder,
		transport:    http.DefaultTransport,
		selector:     wrr.New(),
	}
	for _, o := range opts {
		o(conn)
	}
	if conn.tlsConf != nil {
		if tr, ok := conn.transport.(*http.Transport); ok {
			tr.TLSClientConfig = conn.tlsConf
		}
	}
	insecure := conn.tlsConf == nil
	target, err := newTarget(conn.endpoint, insecure)
	if err != nil {
		return nil, err
	}
	var r *resolver
	if conn.discovery != nil {
		if target.Scheme != "discovery" {
			return nil, fmt.Errorf("[http client] invalid scheme : %v", target.Scheme)
		}
		target.Path = ""
		if r, err = newResolver(ctx, conn.discovery, target, conn.selector, insecure); err != nil {
			return nil, fmt.Errorf("[http client] new resolver failed!err: %v", conn.endpoint)
		}
	}
	conn.target = target
	conn.insecure = insecure
	conn.r = r
	conn.cc = &http.Client{
		Timeout:   conn.timeout,
		Transport: conn.transport,
	}
	conn.streamCc = &http.Client{
		Transport: conn.transport,
	}
	return conn, nil
}

// Invoke makes an rpc call procedure for remote service.
func (conn *ClientConn) Invoke(ctx context.Context, method, reqPath string, args interface{}, reply interface{}, opts ...CallOption) error {
	var (
		contentType string
		body        io.Reader
	)
	c := defaultCallInfo(reqPath)
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return err
		}
	}
	if args != nil {
		data, err := conn.encoder(ctx, c.contentType, args)
		if err != nil {
			return err
		}
		contentType = c.contentType
		body = bytes.NewReader(data)
	}
	url := fmt.Sprintf("%s://%s%s", conn.target.Scheme, conn.target.Authority, path.Join(conn.target.Path, reqPath))
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", c.contentType)
	}
	if conn.userAgent != "" {
		req.Header.Set("User-Agent", conn.userAgent)
	}
	ctx = rpc.NewClientContext(ctx, &Transport{
		endpoint:     conn.endpoint,
		reqHeader:    headerCarrier(req.Header),
		operation:    c.operation,
		request:      req,
		pathTemplate: c.pathTemplate,
	})
	return conn.invoke(ctx, req, args, reply, c, opts...)
}

func (conn *ClientConn) invoke(ctx context.Context, req *http.Request, args interface{}, reply interface{}, c callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		res, err := conn.do(req.WithContext(ctx))
		if res != nil {
			cs := csAttempt{res: res}
			for _, o := range opts {
				o.after(&c, &cs)
			}
		}
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if err := conn.decoder(ctx, res, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}
	if len(conn.middleware) > 0 {
		h = middleware.Chain(conn.middleware...)(h)
	}
	r, err := h(ctx, args)
	if err != nil {
		return err
	}
	if reply != nil && r != nil {
		if r == reply {
			return nil
		}
		return copier.Copy(reply, r)
	}
	return nil
}

// Stream makes an rpc call procedure for remote service.
func (conn *ClientConn) Stream(ctx context.Context, method, reqPath string, args interface{}, opts ...CallOption) (IClientStream, error) {
	var (
		contentType string
		body        io.Reader
	)
	c := defaultCallInfo(reqPath)
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return nil, err
		}
	}
	if args != nil {
		data, err := conn.encoder(ctx, c.contentType, args)
		if err != nil {
			return nil, err
		}
		contentType = c.contentType
		body = bytes.NewReader(data)
	}
	url := fmt.Sprintf("%s://%s%s", conn.target.Scheme, conn.target.Authority, path.Join(conn.target.Path, reqPath))
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", c.contentType)
	}
	if conn.userAgent != "" {
		req.Header.Set("User-Agent", conn.userAgent)
	}
	ctx = rpc.NewClientContext(ctx, &Transport{
		endpoint:     conn.endpoint,
		reqHeader:    headerCarrier(req.Header),
		operation:    c.operation,
		request:      req,
		pathTemplate: c.pathTemplate,
	})
	return conn.stream(ctx, req, args, c, opts...)
}

func (conn *ClientConn) stream(ctx context.Context, req *http.Request, args interface{}, c callInfo, opts ...CallOption) (IClientStream, error) {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		res, err := conn.doStream(req.WithContext(ctx))
		if res != nil {
			cs := csAttempt{res: res}
			for _, o := range opts {
				o.after(&c, &cs)
			}
		}
		if err != nil {
			return nil, err
		}
		stream := NewHttpClientStream(res.Body, CodecForResponse(res))
		return stream, nil
	}
	if len(conn.middleware) > 0 {
		h = middleware.Chain(conn.middleware...)(h)
	}
	stream, err := h(ctx, args)
	if err != nil {
		return nil, err
	}
	return stream.(IClientStream), nil
}

// Do send an HTTP request and decodes the body of response into target.
// returns an error (of type *Error) if the response status code is not 2xx.
func (conn *ClientConn) Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	c := defaultCallInfo(req.URL.Path)
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return nil, err
		}
	}

	return conn.do(req)
}

func (conn *ClientConn) doStream(req *http.Request) (*http.Response, error) {
	var done func(context.Context, selector.DoneInfo)
	if conn.r != nil {
		var (
			err  error
			node selector.Node
		)
		if node, done, err = conn.r.Select(req.Context()); err != nil {
			return nil, err
		}
		if conn.insecure {
			req.URL.Scheme = "http"
		} else {
			req.URL.Scheme = "https"
		}
		req.URL.Host = node.Address()
		req.Host = node.Address()
	}
	resp, err := conn.streamCc.Do(req)
	if err == nil {
		err = conn.errorDecoder(req.Context(), resp)
	}
	if done != nil {
		done(req.Context(), selector.DoneInfo{Err: err})
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (conn *ClientConn) do(req *http.Request) (*http.Response, error) {
	var done func(context.Context, selector.DoneInfo)
	if conn.r != nil {
		var (
			err  error
			node selector.Node
		)
		if node, done, err = conn.r.Select(req.Context()); err != nil {
			return nil, err
		}
		if conn.insecure {
			req.URL.Scheme = "http"
		} else {
			req.URL.Scheme = "https"
		}
		req.URL.Host = node.Address()
		req.Host = node.Address()
	}
	resp, err := conn.cc.Do(req)
	if err == nil {
		err = conn.errorDecoder(req.Context(), resp)
	}
	if done != nil {
		done(req.Context(), selector.DoneInfo{Err: err})
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Close tears down the Transport and all underlying connections.
func (conn *ClientConn) Close() error {
	if conn.r != nil {
		return conn.r.Close()
	}
	return nil
}

// DefaultRequestEncoder is an HTTP request encoder.
func DefaultRequestEncoder(ctx context.Context, contentType string, in interface{}) ([]byte, error) {
	name := httputil.ContentSubtype(contentType)
	body, err := encoding.GetCodec(name).Marshal(in)
	if err != nil {
		return nil, err
	}
	return body, err
}

// DefaultResponseDecoder is an HTTP response decoder.
func DefaultResponseDecoder(ctx context.Context, res *http.Response, v interface{}) error {
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return CodecForResponse(res).Unmarshal(data, v)
}

// DefaultErrorDecoder is an HTTP error decoder.
func DefaultErrorDecoder(ctx context.Context, res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 400 {
		return errors.New(string(data))
	}
	e := make(map[string]interface{})
	err = CodecForResponse(res).Unmarshal(data, &e)
	if err != nil {
		return err
	}
	return errors.New(fmt.Sprint(e["err"]))
}

// CodecForResponse get encoding.Codec via http.Response
func CodecForResponse(r *http.Response) encoding.Codec {
	codec := encoding.GetCodec(httputil.ContentSubtype(r.Header.Get("Content-Type")))
	if codec != nil {
		return codec
	}
	return encoding.GetCodec("json")
}
