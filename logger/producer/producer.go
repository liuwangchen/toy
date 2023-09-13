package producer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

const (
	MaxLogBuffer = 4096
)

type LogProducer interface {
	IsDone() <-chan struct{}
	Log(topic string, key string, val []byte)
	Close() error
}

// KafkaLogProducerConfig 配置
type KafkaLogProducerConfig struct {
	Addr       string `xml:"addr" yaml:"addr"`
	Frequency  int    `xml:"frequency" yaml:"frequency"`
	MaxMessage int    `xml:"max_message" yaml:"max_message"`
}

// KafkaLogProducer kafka日志生产者
type KafkaLogProducer struct {
	producer sarama.AsyncProducer

	msgQ      chan *msg
	wg        sync.WaitGroup
	closeChan chan struct{}
}

type msg struct {
	topic string
	key   string
	val   []byte
}

// NewKafkaLogProducer 构造KafkaProducer
func NewKafkaLogProducer(cfg *KafkaLogProducerConfig) (*KafkaLogProducer, error) {

	kafkaConfig := sarama.NewConfig()
	var (
		clientId   = "log_pruducer"
		frequency  = 500
		maxMessage = MaxLogBuffer
	)
	if cfg.Frequency > 0 {
		frequency = cfg.Frequency
	}
	if cfg.MaxMessage > 0 {
		maxMessage = cfg.MaxMessage
	}
	// sarama.Logger = log.New(os.Stdout, fmt.Sprintf("[%s]", cfg.Name), log.LstdFlags)
	kafkaConfig.ClientID = clientId
	// 等待服务器所有副本都保存成功后的响应
	kafkaConfig.Producer.RequiredAcks = sarama.NoResponse       // Only wait for the leader to ack
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy // Compress messages
	kafkaConfig.ChannelBufferSize = MaxLogBuffer
	// kafkaConfig.KafkaLogProducer.Flush.Bytes = 1024                     // 1MB
	// kafkaConfig.KafkaLogProducer.Flush.Messages = 64
	// kafkaConfig.KafkaLogProducer.Flush.MaxMessages = kafkaConfig.KafkaLogProducer.Flush.Messages * 2
	kafkaConfig.Producer.Flush.Frequency = time.Duration(frequency) * time.Millisecond // Flush batches every 500ms
	//随机的分区类型
	kafkaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	kafkaConfig.Metadata.RefreshFrequency = time.Minute * time.Duration(5)
	kafkaConfig.Net.KeepAlive = time.Second * time.Duration(10)
	kafkaConfig.Metadata.Timeout = time.Second * 10

	p, err := sarama.NewAsyncProducer(strings.Split(cfg.Addr, ","), kafkaConfig)
	if err != nil {
		return nil, err
	}
	ret := &KafkaLogProducer{
		producer:  p,
		msgQ:      make(chan *msg, maxMessage),
		closeChan: make(chan struct{}),
	}

	return ret, nil
}

// Run 运行
func (p *KafkaLogProducer) Run() {

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

	LOOP:
		for {
			select {
			case msg := <-p.msgQ:
				m := &sarama.ProducerMessage{
					Topic: msg.topic,
					Key:   sarama.StringEncoder(msg.key),
					Value: sarama.ByteEncoder(msg.val),
				}
				p.producer.Input() <- m
			case <-p.closeChan:
				break LOOP
			}
		}

		for hasTask := true; hasTask; {
			select {
			case msg := <-p.msgQ:
				m := &sarama.ProducerMessage{
					Topic: msg.topic,
					Key:   sarama.StringEncoder(msg.key),
					Value: sarama.ByteEncoder(msg.val),
				}
				p.producer.Input() <- m
			default:
				hasTask = false
			}
		}

	}()

	go func() {
		for err := range p.producer.Errors() {
			fmt.Printf("[producer]err=[%s] topic=[%s] key=[%s] val=[%s] \n", err.Error(), err.Msg.Topic, err.Msg.Key, err.Msg.Value)
		}
	}()

}

// IsDone() 是否已经结束
func (p *KafkaLogProducer) IsDone() <-chan struct{} {
	return p.closeChan
}

// Close 关闭
func (p *KafkaLogProducer) Close() error {
	close(p.closeChan)
	p.wg.Wait()

	return p.producer.Close()
}

// Log 发送log
func (p *KafkaLogProducer) Log(topic string, key string, val []byte) {
	m := &msg{
		topic: topic,
		key:   key,
		val:   val,
	}
	p.msgQ <- m
}
