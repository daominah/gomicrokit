package kafka

import (
	"strings"

	"time"

	"github.com/Shopify/sarama"
	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/maths"
	"github.com/pkg/errors"
)

var (
	ErWriteTimeout = errors.New("write message timeout")
)

type ProducerConfig struct {
	// comma separated list: broker1:9092,broker2:9092,broker3:9092
	BrokersList  string
	DefaultTopic string
}

type Producer struct {
	defaultTopic string
	samProducer  sarama.AsyncProducer
}

func NewProducer(conf ProducerConfig) (*Producer, error) {
	log.Infof("creating a producer from %#v", conf)
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
			log.Infof("delivered msgId %v to topic %v:%v:%v",
				sent.Metadata, sent.Topic, sent.Partition, sent.Offset)
		}
	}()
	// TODO: check if disconnect
	return p, nil
}

// messages have a same key will be sent to same partition.
func (p Producer) SendExplicitMessage(topic string, value string, key string) error {
	uniqueId := maths.GenUUID()[:8]
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
		log.Infof("sending msgId %v to %v:%v: %v",
			uniqueId, samMsg.Topic, key, samMsg.Value)
		err = nil
	case <-time.After(1 * time.Second):
		err = ErWriteTimeout
	}
	return err
}

func (p Producer) SendMessage(value string) error {
	return p.SendExplicitMessage(p.defaultTopic, value, "")
}
