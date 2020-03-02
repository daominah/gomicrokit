package kafka

import (
	"strings"

	"time"

	"github.com/Shopify/sarama"
	"github.com/daominah/gomicrokit/gofast"
	"github.com/daominah/gomicrokit/log"
	"github.com/pkg/errors"
)

// Errors when produce
var (
	ErrWriteTimeout = errors.New("write message timeout")
)

// ProducerConfig _
type ProducerConfig struct {
	// BrokersList is a comma separated list: "broker1:9092,broker2:9092,broker3:9092"
	BrokersList string
	// DefaultTopic should not be empty
	DefaultTopic string
}

// Producer _
type Producer struct {
	defaultTopic string
	samProducer  sarama.AsyncProducer
}

// NewProducer returns a connected Producer
func NewProducer(conf ProducerConfig) (*Producer, error) {
	log.Infof("creating a producer with %#v", conf)
	// construct sarama config
	samConf := sarama.NewConfig()
	samConf.Producer.RequiredAcks = sarama.WaitForLocal
	samConf.Producer.Retry.Max = 10
	samConf.Producer.Return.Successes = true

	// connect to kafka
	p := &Producer{defaultTopic: conf.DefaultTopic}
	brokers := strings.Split(conf.BrokersList, ",")
	var err error
	p.samProducer, err = sarama.NewAsyncProducer(brokers, samConf)
	if err != nil {
		return nil, errors.Wrap(err, "error when create producer")
	}
	log.Infof("connected to kafka cluster %v", conf.BrokersList)
	go func() {
		for err := range p.samProducer.Errors() {
			errMsg := err.Err.Error()
			if errMsg == "circuit breaker is open" {
				errMsg = "probably you did not assign topic"
			}
			log.Infof("failed to write msgId %v to topic %v: %v",
				err.Msg.Metadata, err.Msg.Topic, errMsg)
		}
	}()
	go func() {
		for sent := range p.samProducer.Successes() {
			log.Condf(LOG, "delivered msgId %v to topic %v:%v:%v",
				sent.Metadata, sent.Topic, sent.Partition, sent.Offset)
		}
	}()
	return p, nil
}

// SendExplicitMessage sends messages have a same key to same partition
func (p Producer) SendExplicitMessage(topic string, value string, key string) error {
	uniqueId := gofast.GenUUID()[:8]
	samMsg := &sarama.ProducerMessage{
		Value:    sarama.StringEncoder(value),
		Topic:    topic,
		Metadata: uniqueId,
	}
	if key != "" {
		samMsg.Key = sarama.StringEncoder(key)
	}
	var err error
	select {
	case p.samProducer.Input() <- samMsg:
		log.Condf(LOG, "sending msgId %v to %v:%v: %v",
			uniqueId, samMsg.Topic, key, samMsg.Value)
		err = nil
	case <-time.After(1 * time.Second):
		err = ErrWriteTimeout
	}
	return err
}

// SendMessage sends message to a random partition of defaultTopic
func (p Producer) SendMessage(value string) error {
	return p.SendExplicitMessage(p.defaultTopic, value, "")
}
