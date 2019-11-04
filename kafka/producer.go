package kafka

import (
	"strings"

	"github.com/Shopify/sarama"
	"github.com/daominah/gomicrokit/log"
	"github.com/pkg/errors"
	"github.com/daominah/gomicrokit/maths"
	"fmt"
)

type ProducerConfig struct {
	// comma separated list: broker1:9092,broker2:9092,broker3:9092
	BrokersList  string
	DefaultTopic string
}

type Producer struct {
	defaultTopic string
	samProducer  sarama.SyncProducer
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
	p.samProducer, err = sarama.NewSyncProducer(brokers, samConf)
	if err != nil {
		return nil, errors.Wrap(err, "error when create producer")
	}
	log.Infof("connected to kafka cluster %v", conf.BrokersList)

	return p, nil
}

// messages have a same key will be sent to same partition.
func (p Producer) SendExplicitMessage(topic string, value string, key string) error {
	uniqueId := maths.GenUUID()[:8]
	samMsg := &sarama.ProducerMessage{
		Value: sarama.StringEncoder(value),
		Topic: topic,
		Key:   sarama.StringEncoder(key),
	}
	log.Infof("sending msg %v to topic %v:%v: %v",
		uniqueId, samMsg.Topic, samMsg.Key, samMsg.Value)
	partition, offset, err := p.samProducer.SendMessage(samMsg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error when send msg %v", uniqueId))
	}
	log.Infof("delivered msg %v to topic %v:%v:%v",
		uniqueId, samMsg.Topic, partition, offset)
	return nil
}

func (p Producer) SendMessage(value string) error {
	return p.SendExplicitMessage(p.defaultTopic, value, "")
}
