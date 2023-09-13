package natsrpc

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/liuwangchen/toy/transport/middleware/trace"
	"github.com/nats-io/nats.go"
)

// WithServerErrorHandler error handler
func WithServerErrorHandler(h func(metadata ServerConnMetadata, err error)) Option {
	return func(s ISetOption) {
		server, ok := s.(*ServerConn)
		if !ok {
			return
		}
		server.errorHandler = h
	}
}

// ServerConn RPC conn
type ServerConn struct {
	address     string
	encType     string
	conn        *nats.Conn
	enc         *nats.EncodedConn // NATS Encode ClientConn
	natsOptions []nats.Option
	mw          []middleware.Middleware // middleware
	namespace   string                  // ns
	timeout     time.Duration

	wg               sync.WaitGroup                               // wait group
	mu               sync.Mutex                                   // lock
	services         map[*service][]*nats.Subscription            // 服务 name->service
	errorHandler     func(metadata ServerConnMetadata, err error) // error handler
	endpoint         *url.URL
	ready            bool
	isSelfCreateConn bool
}

var _ ISetOption = (*ServerConn)(nil)

// NewServerConn 构造器
func NewServerConn(option ...Option) (*ServerConn, error) {
	s := &ServerConn{
		encType:  PROTOBUF_ENCODER,
		services: make(map[*service][]*nats.Subscription),
		errorHandler: func(metadata ServerConnMetadata, err error) {
			logger.ErrorW("ServerConn.handle error", "serviceName", metadata.ServiceName, "methodName", metadata.MethodName, "subject", metadata.MethodSubject, "err", err.Error())
		},
		timeout: time.Duration(3) * time.Second,
	}
	for _, v := range option {
		v(s)
	}
	// 如果未提供conn，则创建一个conn
	if s.conn == nil {
		conn, err := nats.Connect(s.address, s.natsOptions...)
		if err != nil {
			return nil, err
		}
		s.conn = conn
		s.isSelfCreateConn = true
	}

	clientID, err := s.conn.GetClientID()
	if err != nil {
		return nil, err
	}

	s.endpoint = endpoint.NewEndpoint("nats", fmt.Sprintf("%d@%s", clientID, s.conn.ConnectedAddr()), false)
	enc, err := nats.NewEncodedConn(s.conn, s.encType)
	if err != nil {
		return nil, err
	}
	s.enc = enc
	return s, nil
}

func (s *ServerConn) SetConn(conn *nats.Conn) {
	s.conn = conn
}

func (s *ServerConn) SetEncodeType(encType string) {
	s.encType = encType
}

func (s *ServerConn) SetAddr(addr string) {
	s.address = addr
}

func (s *ServerConn) SetNatsOption(natsOptions ...nats.Option) {
	s.natsOptions = natsOptions
}

func (s *ServerConn) SetConnMiddleWare(mw ...middleware.Middleware) {
	s.mw = mw
}

func (s *ServerConn) SetNamespace(namespace string) {
	s.namespace = namespace
}

func (s *ServerConn) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// Close 关闭
func (s *ServerConn) Close(ctx context.Context) (err error) {
	s.ClearAllSubscription()

	over := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(over)
	}()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-over:
	}
	if err1 := s.enc.Flush(); err == nil && err1 != nil {
		err = err1
	}
	if s.isSelfCreateConn {
		s.conn.Close()
	}
	return
}

// ClearAllSubscription 取消所有订阅
func (s *ServerConn) ClearAllSubscription() {
	s.mu.Lock()
	ss := make([]*service, 0, len(s.services))
	for s := range s.services {
		ss = append(ss, s)
	}
	s.mu.Unlock()

	for _, v := range ss {
		s.remove(v)
	}
}

// Unregister 反注册
func (s *ServerConn) remove(service *service) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.services[service]
	if ok {
		for _, v := range sub {
			v.Unsubscribe()
		}
		delete(s.services, service)
	}
	return ok
}

// Register 注册服务
func (s *ServerConn) Register(name string, svc interface{}, opts ...ServiceOption) error {
	// new 一个服务
	serv, err := newService(name, svc, opts...)
	if nil != err {
		return err
	}
	if serv.timeout == 0 {
		serv.timeout = s.timeout
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否重复
	if _, ok := s.services[serv]; ok {
		return fmt.Errorf("service [%s] duplicate", serv.serviceName)
	}
	serv.conn = s
	s.services[serv] = make([]*nats.Subscription, 0)
	return nil
}

// subscribeMethod 订阅服务的方法
func (s *ServerConn) subscribeMethod(service *service) error {
	dup := make(map[string]struct{}, len(service.methods))
	// 订阅
	for methodName, v := range service.methods {
		m := v
		methodNameTemp := methodName
		reqRspIds, ok := service.methodReqRspIds[methodName]
		if ok {
			methodNameTemp = fmt.Sprint(reqRspIds[0])
		}
		methodSub := CombineStr(s.namespace, service.serviceName, methodNameTemp)

		// 主题
		subject := CombineStr(methodSub, service.topic)

		// 重复sub判断
		_, ok = dup[subject]
		if ok {
			return fmt.Errorf("dup subject %s", subject)
		}
		dup[subject] = struct{}{}

		// queue，优先用方法级别的queue，没有的话在用service级别的queue
		queue, ok := service.methodQueue[methodName]
		if !ok {
			queue = service.queue
		}

		// 是否顺序处理msg
		isSequenceHandleMsg := service.methodSequenceHandle[methodName]

		// 订阅
		natsSub, subErr := s.enc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			cb := func() {
				serverConnMetadata := ServerConnMetadata{
					Namespace:         s.namespace,
					ServiceName:       service.serviceName,
					ServiceSimpleName: service.serviceName,
					MethodSubject:     methodSub,
					IsPublish:         m.isPublish,
					ReqType:           m.reqType,
					RspType:           m.respType,
					ReplySub:          msg.Reply,
					Encoder:           s.enc.Enc,
					MethodName:        m.name,
					ReqRspIds:         reqRspIds,
				}
				ctx := s.withServerMetadata(context.Background(), serverConnMetadata)
				err := s.handle(ctx, service, m, msg)
				if err != nil {
					s.errorHandler(serverConnMetadata, err)
				}
			}
			if isSequenceHandleMsg {
				cb()
			} else {
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					cb()
				}()
			}
		})
		if nil != subErr {
			return subErr
		}
		s.services[service] = append(s.services[service], natsSub)
	}
	return nil
}

func (s *ServerConn) handle(ctx context.Context, service *service, m *method, msg *nats.Msg) error {
	req := m.newRequest()

	if len(msg.Data) > 0 {
		rpcReq := &Request{}
		if err := s.enc.Enc.Decode(msg.Subject, msg.Data, rpcReq); nil != err {
			return err
		}
		if len(rpcReq.Header) > 0 {
			// 包traceId
			ctx = trace.ContextWithTraceId(ctx, trace.GetTraceIdFromHeader(rpcReq.Header))
			// 包header
			ctx = WithHeaderContext(ctx, rpcReq.Header)
		}
		if len(rpcReq.Payload) > 0 {
			if err := s.enc.Enc.Decode(msg.Subject, rpcReq.Payload, req); nil != err {
				return err
			}
		}
	}

	var (
		reply interface{}
		err   error
	)
	// handle
	reply, err = service.handle(ctx, m, req)
	// publish的情况
	if len(msg.Reply) == 0 || m.isPublish {
		if len(msg.Reply) > 0 {
			msg.Ack()
		}
		return err
	}
	rp := &Reply{}
	if err != nil {
		rp.Error = err.Error()
	} else {
		b, err := s.enc.Enc.Encode(msg.Subject, reply)
		if err != nil {
			return err
		}
		rp.Payload = b
	}
	b, err := s.enc.Enc.Encode(msg.Subject, rp)
	if err != nil {
		return err
	}
	if s.enc.Conn.IsClosed() {
		return fmt.Errorf("conn colsed")
	}

	// 注入replyHeader
	replyHeader := map[string][]string{}
	service.injectMethodReqRespIdsIntoHeader(m, replyHeader)

	// 构造恢复msg
	respMsg := &nats.Msg{
		Subject: msg.Reply,
		Data:    b,
		Header:  replyHeader,
	}
	return s.enc.Conn.PublishMsg(respMsg)
}

// Start 运行
func (s *ServerConn) Start(ctx context.Context) error {
	for serv := range s.services {
		if err := s.subscribeMethod(serv); nil != err {
			return err
		}
	}
	s.ready = true
	addr := s.address
	if len(addr) == 0 {
		addr = s.conn.ConnectedAddr()
	}
	logger.Info("[Nats] server listening on: %s", addr)
	<-ctx.Done()
	return nil
}

func (s *ServerConn) Endpoint() (*url.URL, error) {
	return s.endpoint, nil
}

// Stop 停止
func (s *ServerConn) Stop(ctx context.Context) error {
	logger.Info("[Nats] server stopping")
	return s.Close(ctx)
}

type ServerConnMetadata struct {
	Namespace         string
	ServiceName       string
	ServiceSimpleName string
	MethodSubject     string
	IsPublish         bool
	ReqType           reflect.Type
	RspType           reflect.Type

	ReplySub   string
	Encoder    nats.Encoder
	MethodName string
	ReqRspIds  [2]int32
}

type serverMetadataKey struct{}

// withServerMetadata 填充客户端元数据
func (s *ServerConn) withServerMetadata(ctx context.Context, metadata ServerConnMetadata) context.Context {
	newCtx := context.WithValue(ctx, serverMetadataKey{}, metadata)
	return newCtx
}

// ServerMetadataFromCtx 获得客户端元数据
func ServerMetadataFromCtx(ctx context.Context) ServerConnMetadata {
	if ctx == nil {
		return ServerConnMetadata{}
	}
	val := ctx.Value(serverMetadataKey{})
	if val != nil {
		return val.(ServerConnMetadata)
	}
	return ServerConnMetadata{}
}

func (s *ServerConn) GetConn() *nats.Conn {
	return s.conn
}

func (s *ServerConn) Ready() bool {
	return s.ready
}
