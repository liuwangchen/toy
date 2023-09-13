package logger

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/liuwangchen/toy/logger/producer"
)

// ProducerLogWriter 生产者Logger
type ProducerLogWriter struct {
	topic    string
	key      string
	clientIp string

	p producer.LogProducer

	msgQ chan *LogRecord
}

// NewProducerLogWriter 构造
func NewProducerLogWriter(topic string, key string, clientIp string, p producer.LogProducer) *ProducerLogWriter {
	l := &ProducerLogWriter{
		msgQ:     make(chan *LogRecord, producer.MaxLogBuffer),
		p:        p,
		topic:    topic,
		key:      key,
		clientIp: clientIp,
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go l.run()
	}
	return l
}

func (w *ProducerLogWriter) AddFilter(lvl string) {
	if "FINEST" == lvl {
		Global.AddFilter("kafka", FINEST, w)
	} else if "FINE" == lvl {
		Global.AddFilter("kafka", FINE, w)
	} else if "DEBUG" == lvl {
		Global.AddFilter("kafka", DEBUG, w)
	} else if "TRACE" == lvl {
		Global.AddFilter("kafka", TRACE, w)
	} else if "INFO" == lvl {
		Global.AddFilter("kafka", INFO, w)
	} else if "WARNING" == lvl {
		Global.AddFilter("kafka", WARNING, w)
	} else if "ERROR" == lvl {
		Global.AddFilter("kafka", ERROR, w)
	} else if "FATAL" == lvl {
		Global.AddFilter("kafka", FATAL, w)
	} else {
		Global.AddFilter("kafka", WARNING, w)
	}

}

func (w *ProducerLogWriter) run() {

	for {
		select {
		case rec, ok := <-w.msgQ:
			if !ok {
				return
			}
			buff := &bytes.Buffer{}
			buff.WriteString("[")
			buff.WriteString(rec.Created.Format("2006/01/02 15:04:05.000000 -0700"))
			buff.WriteString("]")
			buff.WriteString(" ")

			buff.WriteString(w.clientIp)
			buff.WriteString(" ")

			buff.WriteString("[")
			buff.WriteString(rec.Level.String())
			buff.WriteString("]")
			buff.WriteString(" ")

			buff.WriteString("(")
			buff.WriteString(rec.Source)
			buff.WriteString(")")
			buff.WriteString(" ")

			buff.WriteString(rec.Message)
			w.p.Log(w.topic, w.key, buff.Bytes())
		case <-w.p.IsDone():
			return
		}
	}
}

// LogWrite
func (w *ProducerLogWriter) LogWrite(rec *LogRecord) {
	select {
	case w.msgQ <- rec:
	default:
		fmt.Printf("[ProducerLogWriter] msgQ is full, msgQ.len:%d\n", len(w.msgQ))
	}
}

// Close
func (w *ProducerLogWriter) Close() {
	w.p.Close()
	close(w.msgQ)
}

func toProducerLogWriter(producerType string, props []LogProperty) (*ProducerLogWriter, error) {
	var (
		p     producer.LogProducer
		topic string
		key   string
	)
	switch producerType {
	case "kafka":
		config := &producer.KafkaLogProducerConfig{}
		// Parse properties
		for _, prop := range props {
			switch prop.Name {
			case "addr":
				config.Addr = prop.Value
			case "max_message":
				v, err := strconv.Atoi(prop.Value)
				if err != nil {
					return nil, err
				}
				config.MaxMessage = v
			case "frequency":
				v, err := strconv.Atoi(prop.Value)
				if err != nil {
					return nil, err
				}
				config.Frequency = v
			case "topic":
				topic = prop.Value
			case "key":
				key = prop.Value
			default:
				return nil, fmt.Errorf("LoadConfiguration: Warning: Unknown property \"%s\" for file filter\n", prop.Name)
			}
		}

		kafkaLogProducer, err := producer.NewKafkaLogProducer(config)
		if err != nil {
			return nil, err
		}
		kafkaLogProducer.Run()
		p = kafkaLogProducer
	}
	return NewProducerLogWriter(topic, key, getClientIp(), p), nil
}

// GetLocalIP 获得内网IP
func getClientIp() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return ""
}
