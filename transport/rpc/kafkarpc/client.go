package kafkarpc

import (
	"context"
	"strings"

	"github.com/liuwangchen/toy/transport/encoding"
	"github.com/liuwangchen/toy/transport/middleware"
)

// ClientOption is HTTP client option.
type ClientOption func(*Client)

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

func WithCodec(codec encoding.Codec) ClientOption {
	return func(c *Client) {
		c.codec = codec
	}
}

func WithNamespace(namespace string) ClientOption {
	return func(c *Client) {
		c.namespace = namespace
	}
}

// WithMiddleware with service middleware option.
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(c *Client) {
		c.middlewares = m
	}
}

type Client struct {
	k           *KafkaClient
	endpoint    string
	ctx         context.Context
	codec       encoding.Codec
	namespace   string
	middlewares []middleware.Middleware
}

func (c Client) Publish(ctx context.Context, topic string, msg interface{}) error {
	h := func(ctx context.Context, m interface{}) (interface{}, error) {
		b, err := c.codec.Marshal(m)
		if err != nil {
			return nil, err
		}
		return nil, c.k.Publish(c.Topic(topic), b)
	}
	if len(c.middlewares) > 0 {
		h = middleware.Chain(c.middlewares...)(h)
	}
	_, err := h(ctx, msg)
	return err
}

func (c Client) Topic(topic string) string {
	if len(c.namespace) > 0 {
		topic = strings.Join([]string{c.namespace, topic}, ".")
	}
	return topic
}

func (c Client) Close() error {
	return c.k.Disconnect()
}

func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	c := &Client{
		ctx:   ctx,
		codec: encoding.GetCodec("json"),
	}
	for _, o := range opts {
		o(c)
	}
	c.k = NewKafkaClient(c.ctx, c.endpoint)
	err := c.k.Connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}
