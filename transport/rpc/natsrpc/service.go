package natsrpc

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"time"

	"github.com/liuwangchen/toy/transport/middleware"
	"github.com/nats-io/nats.go"
)

// ServiceOption service option
type ServiceOption func(options *service)

// WithServiceTimeout 超时时间
func WithServiceTimeout(timeout time.Duration) ServiceOption {
	return func(s *service) {
		s.timeout = timeout
	}
}

// WithServiceMiddleware 超时时间
func WithServiceMiddleware(mw ...middleware.Middleware) ServiceOption {
	return func(s *service) {
		s.mw = mw
	}
}

// WithServiceTopic 主题
func WithServiceTopic(topic interface{}) ServiceOption {
	return func(s *service) {
		s.topic = fmt.Sprint(topic)
	}
}

// WithServiceQueue 服务级别的queue
func WithServiceQueue(queue string) ServiceOption {
	return func(s *service) {
		s.queue = queue
	}
}

// WithServiceMethodQueue 方法级别的queue
func WithServiceMethodQueue(methodQueues map[string]string) ServiceOption {
	return func(s *service) {
		s.methodQueue = methodQueues
	}
}

// WithServiceMethodReqRspIds 方法Id
func WithServiceMethodReqRspIds(methodReqRspIds map[string][2]int32) ServiceOption {
	return func(s *service) {
		s.methodReqRspIds = methodReqRspIds
	}
}

// WithServiceMethodSequence  方法顺序处理msg
func WithServiceMethodSequence(methodSequenceHandle map[string]bool) ServiceOption {
	return func(s *service) {
		s.methodSequenceHandle = methodSequenceHandle
	}
}

// service 服务
type service struct {
	serviceName          string                  // 名字
	val                  interface{}             // 值
	conn                 *ServerConn             // rpc
	methods              map[string]*method      // 方法集合
	timeout              time.Duration           // 请求/handle的超时
	mw                   []middleware.Middleware // middleware
	topic                string                  // 主题
	queue                string                  // 服务级别queue
	methodQueue          map[string]string       // 方法级queue
	methodReqRspIds      map[string][2]int32     // 方法reqRspIds
	methodSequenceHandle map[string]bool         // 方法是否顺序处理msg methodName -> bool
}

// Name 名字
func (s *service) Name() string {
	return s.serviceName
}

// Close 关闭
// 会取消所有订阅
func (s *service) Close() bool {
	return s.conn.remove(s)
}

// newService 创建服务
func newService(serviceName string, i interface{}, opts ...ServiceOption) (*service, error) {
	s := &service{
		methods:              map[string]*method{},
		serviceName:          serviceName, // name = package.service
		val:                  i,
		queue:                "default",
		methodSequenceHandle: make(map[string]bool),
		methodReqRspIds:      map[string][2]int32{},
		methodQueue:          map[string]string{},
	}

	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("service must be a pointer")
	}
	typ := reflect.Indirect(val).Type()
	if !ast.IsExported(typ.Name()) {
		return nil, fmt.Errorf("service [%s] must be exported", serviceName)
	}

	ms := parseMethod(i)
	if len(ms) == 0 {
		return nil, fmt.Errorf("service [%s] has no exported method", serviceName)
	}

	for _, v := range ms {
		if _, ok := s.methods[v.name]; ok {
			return nil, fmt.Errorf("service [%s] duplicate method [%s]", serviceName, v.name)
		}
		s.methods[v.name] = v
	}

	for _, v := range opts {
		v(s)
	}
	return s, nil
}

func (s *service) handle(ctx context.Context, m *method, req interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	h := func(ctx context.Context, req1 interface{}) (interface{}, error) {
		return m.handle(s.val, ctx, req1)
	}

	// 中间件
	mw := append(s.conn.mw, s.mw...)
	if len(mw) > 0 {
		h = middleware.Chain(mw...)(h)
	}

	return h(ctx, req)
}

var (
	headerKey_MethodReqId  = "headerMethod_ReqId"
	headerKey_MethodRespId = "headerMethod_RespId"
)

func (s *service) injectMethodReqRespIdsIntoHeader(m *method, header nats.Header) {
	reqRspIds, ok := s.methodReqRspIds[m.name]
	if !ok {
		return
	}
	header.Set(headerKey_MethodReqId, fmt.Sprint(reqRspIds[0]))
	header.Set(headerKey_MethodRespId, fmt.Sprint(reqRspIds[1]))
}

func GetReqIdFromReplyHeader(header nats.Header) int32 {
	v := header.Get(headerKey_MethodReqId)
	reqId, _ := strconv.Atoi(v)
	return int32(reqId)
}

func GetRespIdFromReplyHeader(header nats.Header) int32 {
	v := header.Get(headerKey_MethodRespId)
	respId, _ := strconv.Atoi(v)
	return int32(respId)
}
