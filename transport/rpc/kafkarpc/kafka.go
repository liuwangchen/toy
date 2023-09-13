package kafkarpc

import (
	"context"
	"errors"
	"sync"

	"github.com/Shopify/sarama"
	l4g "github.com/liuwangchen/toy/logger"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

type KafkaClient struct {
	ctx  context.Context
	addr []string

	c  sarama.Client
	p  sarama.SyncProducer
	ap sarama.AsyncProducer

	sc []sarama.Client

	connected bool
	scMutex   sync.Mutex
}

func NewKafkaClient(ctx context.Context, addrs ...string) *KafkaClient {
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:9092"}
	}

	return &KafkaClient{
		ctx:  ctx,
		addr: addrs,
	}
}

func (k *KafkaClient) Address() string {
	if len(k.addr) > 0 {
		return k.addr[0]
	}
	return "127.0.0.1:9092"
}

func (k *KafkaClient) Connect() error {
	if k.connected {
		return nil
	}

	k.scMutex.Lock()
	if k.c != nil {
		k.scMutex.Unlock()
		return nil
	}
	k.scMutex.Unlock()

	pconfig := k.getBrokerConfig()
	// For implementation reasons, the SyncProducer requires
	// `Producer.Return.Errors` and `Producer.Return.Successes`
	// to be set to true in its configuration.
	pconfig.Producer.Return.Successes = true
	pconfig.Producer.Return.Errors = true

	c, err := sarama.NewClient(k.addr, pconfig)
	if err != nil {
		return err
	}

	var (
		ap                   sarama.AsyncProducer
		p                    sarama.SyncProducer
		errChan, successChan = k.getAsyncProduceChan()
	)

	// Because error chan must require, so only error chan
	// If set the error chan, will use async produce
	// else use sync produce
	// only keep one client resource, is c variable
	if errChan != nil {
		ap, err = sarama.NewAsyncProducerFromClient(c)
		if err != nil {
			return err
		}
		// When the ap closed, the Errors() & Successes() channel will be closed
		// So the goroutine will auto exit
		go func() {
			for v := range ap.Errors() {
				errChan <- v
			}
		}()

		if successChan != nil {
			go func() {
				for v := range ap.Successes() {
					successChan <- v
				}
			}()
		}
	} else {
		p, err = sarama.NewSyncProducerFromClient(c)
		if err != nil {
			return err
		}
	}

	k.scMutex.Lock()
	k.c = c
	if p != nil {
		k.p = p
	}
	if ap != nil {
		k.ap = ap
	}
	k.sc = make([]sarama.Client, 0)
	k.connected = true
	k.scMutex.Unlock()

	return nil
}

func (k *KafkaClient) Disconnect() error {
	k.scMutex.Lock()
	defer k.scMutex.Unlock()
	for _, client := range k.sc {
		client.Close()
	}
	k.sc = nil
	if k.p != nil {
		k.p.Close()
	}
	if k.ap != nil {
		k.ap.Close()
	}
	if err := k.c.Close(); err != nil {
		return err
	}
	k.connected = false
	return nil
}

func (k *KafkaClient) Publish(topic string, msg []byte) error {
	var produceMsg = &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
	}

	if k.ap != nil {
		k.ap.Input() <- produceMsg
		return nil
	} else if k.p != nil {
		_, _, err := k.p.SendMessage(produceMsg)
		return err
	}

	return errors.New(`no connection resources available`)
}

var (
	DefaultBrokerConfig  = sarama.NewConfig()
	DefaultClusterConfig = sarama.NewConfig()
)

type brokerConfigKey struct{}
type clusterConfigKey struct{}

type asyncProduceErrorKey struct{}
type asyncProduceSuccessKey struct{}

func (k *KafkaClient) createSaramaClusterClient() (sarama.Client, error) {
	config := k.getClusterConfig()
	cs, err := sarama.NewClient(k.addr, config)
	if err != nil {
		return nil, err
	}
	k.scMutex.Lock()
	defer k.scMutex.Unlock()
	k.sc = append(k.sc, cs)
	return cs, nil
}

func (k *KafkaClient) Subscribe(topic string, handler func(ctx context.Context, msg []byte) error, queue string, autoAck bool) error {
	// we need to create a new client per consumer
	c, err := k.createSaramaClusterClient()
	if err != nil {
		return err
	}
	cg, err := sarama.NewConsumerGroupFromClient(queue, c)
	if err != nil {
		return err
	}

	h := &consumerGroupHandler{
		ctx:     k.ctx,
		handler: handler,
		autoAck: autoAck,
	}
	ctx := context.Background()
	topics := []string{topic}
	go func() {
		for {
			select {
			case err := <-cg.Errors():
				if err != nil {
					l4g.Error("consumer error:", err)
				}
			default:
				err := cg.Consume(ctx, topics, h)
				switch err {
				case sarama.ErrClosedConsumerGroup:
					return
				case nil:
					continue
				default:
					l4g.Error(err)
				}
			}
		}
	}()
	return nil
}

func (k *KafkaClient) String() string {
	return "kafka"
}

func (k *KafkaClient) getBrokerConfig() *sarama.Config {
	if c, ok := k.ctx.Value(brokerConfigKey{}).(*sarama.Config); ok {
		return c
	}
	return DefaultBrokerConfig
}

func (k *KafkaClient) getAsyncProduceChan() (chan<- *sarama.ProducerError, chan<- *sarama.ProducerMessage) {
	var (
		errors    chan<- *sarama.ProducerError
		successes chan<- *sarama.ProducerMessage
	)
	if c, ok := k.ctx.Value(asyncProduceErrorKey{}).(chan<- *sarama.ProducerError); ok {
		errors = c
	}
	if c, ok := k.ctx.Value(asyncProduceSuccessKey{}).(chan<- *sarama.ProducerMessage); ok {
		successes = c
	}
	return errors, successes
}

func (k *KafkaClient) getClusterConfig() *sarama.Config {
	if c, ok := k.ctx.Value(clusterConfigKey{}).(*sarama.Config); ok {
		return c
	}
	clusterConfig := DefaultClusterConfig
	// the oldest supported version is V0_10_2_0
	if !clusterConfig.Version.IsAtLeast(sarama.V0_10_2_0) {
		clusterConfig.Version = sarama.V0_10_2_0
	}
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	return clusterConfig
}

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	ctx     context.Context
	handler func(ctx context.Context, msg []byte) error
	autoAck bool
}

func (*consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.handler(h.ctx, msg.Value)
		if err == nil && h.autoAck {
			sess.MarkMessage(msg, "")
		} else if err != nil {
			l4g.Error("[kafka]: subscriber error: %v", err)
		}
	}
	return nil
}
