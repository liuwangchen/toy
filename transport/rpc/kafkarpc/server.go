package kafkarpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/transport/encoding"
	"github.com/liuwangchen/toy/transport/middleware"
)

type ServerOption func(*Server)

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Middleware with service middleware option.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.middlewares = m
	}
}

// Codec codec
func Codec(codec encoding.Codec) ServerOption {
	return func(s *Server) {
		s.codec = codec
	}
}

// Namespace 命名空间
func Namespace(namespace string) ServerOption {
	return func(s *Server) {
		s.namespace = namespace
	}
}

type Server struct {
	k           *KafkaClient
	address     string
	middlewares []middleware.Middleware // 中间件
	codec       encoding.Codec
	namespace   string
	subs        map[string]*subInfo
	ready       bool
}

type subInfo struct {
	topic   string
	handle  func(ctx context.Context, b []byte) error
	autoAck bool
	queue   string
}

func NewServer(opts ...ServerOption) (*Server, error) {
	s := &Server{
		codec: encoding.GetCodec("json"),
		subs:  map[string]*subInfo{},
	}
	for _, opt := range opts {
		opt(s)
	}
	k := NewKafkaClient(context.Background(), s.address)
	err := k.Connect()
	if err != nil {
		return nil, err
	}
	s.k = k
	return s, nil
}

func (s *Server) GetMiddlewares() []middleware.Middleware {
	return s.middlewares
}

func (s *Server) GetCodec() encoding.Codec {
	return s.codec
}

// Subscribe 订阅
func (s *Server) Subscribe(topic string, handler func(ctx context.Context, b []byte) error, autoAck bool, queue string) {
	newTopic := s.Topic(topic)
	if len(queue) == 0 {
		queue = uuid.New().String()
	}
	s.subs[newTopic] = &subInfo{
		topic:   topic,
		handle:  handler,
		autoAck: autoAck,
		queue:   queue,
	}
}

func (s *Server) Topic(topic string) string {
	if len(s.namespace) > 0 {
		topic = strings.Join([]string{s.namespace, topic}, ".")
	}
	return topic
}

func (s *Server) Start(ctx context.Context) error {
	for _, sub := range s.subs {
		err := s.k.Subscribe(sub.topic, sub.handle, sub.queue, sub.autoAck)
		if err != nil {
			return err
		}
	}
	s.ready = true
	logger.Info("[Kafka] server listening on: %s", s.address)
	<-ctx.Done()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.k == nil {
		return nil
	}
	logger.Info("[Kafka] server stopping")
	return s.k.Disconnect()
}

func (s *Server) Ready() bool {
	return s.ready
}
