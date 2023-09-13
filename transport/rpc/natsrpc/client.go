package natsrpc

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/liuwangchen/toy/pkg/copier"
	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/middleware/trace"
	"github.com/nats-io/nats.go"
)

var _ ISetOption = (*ClientConn)(nil)

// ClientConn RPC client
type ClientConn struct {
	address     string
	encType     string
	conn        *nats.Conn
	enc         *nats.EncodedConn // NATS Encode ClientConn
	natsOptions []nats.Option
	mw          []middleware.Middleware // middleware
	namespace   string                  // ns
	timeout     time.Duration
}

// NewClientConn 构造器
func NewClientConn(opts ...Option) (*ClientConn, error) {
	c := &ClientConn{
		encType: PROTOBUF_ENCODER,
		timeout: time.Duration(3) * time.Second,
	}
	for _, v := range opts {
		v(c)
	}
	// 如果未提供conn，则创建一个conn
	if c.conn == nil {
		conn, err := nats.Connect(c.address)
		if err != nil {
			return nil, err
		}
		c.conn = conn
	}
	encodedConn, err := nats.NewEncodedConn(c.conn, c.encType)
	if err != nil {
		return nil, err
	}
	c.enc = encodedConn
	return c, nil
}

func (c *ClientConn) SetAddr(addr string) {
	c.address = addr
}

func (c *ClientConn) SetConn(conn *nats.Conn) {
	c.conn = conn
}

func (c *ClientConn) SetEncodeType(encType string) {
	c.encType = encType
}

func (c *ClientConn) SetNatsOption(natsOptions ...nats.Option) {
	c.natsOptions = natsOptions
}

func (c *ClientConn) SetConnMiddleWare(mw ...middleware.Middleware) {
	c.mw = mw
}

func (c *ClientConn) SetNamespace(namespace string) {
	c.namespace = namespace
}

func (c *ClientConn) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// ClientOption is an HTTP server option.
type ClientOption func(client *Client)

// WithClientMiddleware with client mw
func WithClientMiddleware(mw ...middleware.Middleware) ClientOption {
	return func(c *Client) {
		c.mw = mw
	}
}

type Client struct {
	conn *ClientConn
	mw   []middleware.Middleware // middleware
}

func NewClient(clientConn *ClientConn, opts ...ClientOption) *Client {
	c := &Client{conn: clientConn}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Publish 发布
func (c *Client) Publish(ctx context.Context, service, method string, req interface{}) error {
	return c.call(ctx, service, method, req, nil)
}

// Request 求
func (c *Client) Request(ctx context.Context, service, method string, req interface{}, rep interface{}) error {
	return c.call(ctx, service, method, req, rep)
}

func (c *Client) call(ctx context.Context, serviceName, methodName string, req interface{}, rep interface{}) error {
	// timeout
	if c.conn.timeout > 0 {
		newCtx, cancel := context.WithTimeout(ctx, c.conn.timeout)
		defer cancel()
		ctx = newCtx
	}

	isPublish := rep == nil

	// subject
	subject := CombineStr(c.conn.namespace, serviceName, methodName)

	// metadata
	splits := strings.Split(serviceName, ".")
	clientConnMetadata := ClientConnMetadata{
		Namespace:     c.conn.namespace,
		ServiceName:   serviceName,
		MethodName:    methodName,
		IsPublish:     isPublish,
		MethodSubject: subject,
		ReqType:       reflect.TypeOf(req),
		RspType:       reflect.TypeOf(rep),
		Encoder:       c.conn.enc.Enc,
	}
	if len(splits) > 0 {
		clientConnMetadata.ServiceSimpleName = splits[len(splits)-1]
	}
	ctx = c.withClientMetadata(ctx, clientConnMetadata)

	ctx = WithHeaderContext(ctx, map[string]string{})

	h := func(ctx1 context.Context, req1 interface{}) (interface{}, error) {
		// 取header
		header := HeaderFromCtx(ctx1)

		// traceId
		trace.SetTraceIdIntoHeader(header, trace.GetTraceIdFromCtx(ctx))

		// 取动态topic
		callTopic := CallTopicFromCtx(ctx1)

		// 最终subject
		subject = CombineStr(subject, callTopic)

		// req
		rpcReq, err := NewRequestWithEncoder(subject, req1, header, c.conn.enc.Enc)
		if err != nil {
			return nil, err
		}
		if isPublish { // publish
			return nil, c.conn.enc.Publish(subject, rpcReq)
		} else { // request
			rp := &Reply{}
			// call
			err = c.conn.enc.RequestWithContext(ctx1, subject, rpcReq, rp)
			if err != nil {
				return nil, err
			}
			if len(rp.Error) > 0 {
				return nil, errors.New(rp.Error)
			}
			// decode
			if err := c.conn.enc.Enc.Decode(subject, rp.Payload, rep); err != nil {
				return nil, err
			}
			return rep, nil
		}
	}

	// 中间件
	mw := append(c.conn.mw[:], c.mw...)
	if len(mw) > 0 {
		h = middleware.Chain(mw...)(h)
	}

	r, err := h(ctx, req)
	if err != nil {
		return err
	}
	if rep != nil && r != nil {
		if r == rep {
			return nil
		}
		return copier.Copy(rep, r)
	}
	return nil
}

// WaitForRsp 等待回复sub后反序列化rsp
func WaitForRsp(ctx context.Context, conn *nats.Conn, replySub string, enc nats.Encoder, rsp interface{}) error {
	// 等待reply
	msg, err := WaitForMsg(ctx, conn, replySub)
	if err != nil {
		return err
	}

	// 反序列化natsrpc
	rp := new(Reply)
	err = enc.Decode(replySub, msg.Data, rp)
	if err != nil {
		return err
	}
	if len(rp.Error) > 0 {
		return errors.New(rp.Error)
	}

	// 反序列化rsp
	err = enc.Decode(replySub, rp.Payload, rsp)
	if err != nil {
		return err
	}
	return nil
}

// WaitForSubscribeRsp 等待回复sub后反序列化rsp
func WaitForSubscribeRsp(ctx context.Context, sub *nats.Subscription, enc nats.Encoder, rsp interface{}) error {
	// 等待reply
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return err
	}

	// 反序列化natsrpc
	rp := new(Reply)
	err = enc.Decode(sub.Subject, msg.Data, rp)
	if err != nil {
		return err
	}
	if len(rp.Error) > 0 {
		return errors.New(rp.Error)
	}

	// 反序列化rsp
	err = enc.Decode(sub.Subject, rp.Payload, rsp)
	if err != nil {
		return err
	}
	return nil
}

func WaitForMsg(ctx context.Context, conn *nats.Conn, replySub string) (*nats.Msg, error) {
	// 等待reply
	sync, err := conn.SubscribeSync(replySub)
	if err != nil {
		return nil, err
	}
	defer sync.Unsubscribe()
	return sync.NextMsgWithContext(ctx)
}

func NewRequestWithEncoder(subject string, req interface{}, header map[string]string, encoder nats.Encoder) (*Request, error) {
	payload, err := encoder.Encode(subject, req)
	if err != nil {
		return nil, err
	}
	return NewRequest(payload, header), nil
}

func NewRequest(payload []byte, header map[string]string) *Request {
	return &Request{
		Payload: payload,
		Header:  header,
	}
}

type ClientConnMetadata struct {
	Namespace         string
	ServiceName       string
	ServiceSimpleName string
	MethodName        string
	MethodSubject     string
	IsPublish         bool
	ReqType           reflect.Type
	RspType           reflect.Type
	Encoder           nats.Encoder
}

type clientMetadataKey struct{}

// withClientMetadata 填充客户端元数据
func (c *Client) withClientMetadata(ctx context.Context, metadata ClientConnMetadata) context.Context {
	newCtx := context.WithValue(ctx, clientMetadataKey{}, metadata)
	return newCtx
}

// ClientMetadataFromCtx 获得客户端元数据
func ClientMetadataFromCtx(ctx context.Context) ClientConnMetadata {
	if ctx == nil {
		return ClientConnMetadata{}
	}
	val := ctx.Value(clientMetadataKey{})
	if val == nil {
		return ClientConnMetadata{}
	}
	return val.(ClientConnMetadata)
}

func (c *ClientConn) GetConn() *nats.Conn {
	return c.conn
}
