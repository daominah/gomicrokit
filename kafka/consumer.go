package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type Offset int

const (
	Earliest = -1
	Latest   = -2
)

// ConsumerConfig should be created by NewConsumerConfig (for default values)
type ConsumerConfig struct {
	// broker1:9092,broker2:9092,broker3:9092
	BootstrapServer string
	Topic           string
	GroupId         string
	Offset          Offset
}

func NewConsumerConfig(
	brokersUrl string, topic string, groupId string, offset Offset) *ConsumerConfig {
	conf := &ConsumerConfig{
		BootstrapServer: "127.0.0.1:9092",
		Topic:           "topic0",
		GroupId:         "",
		Offset:          Earliest,
	}
	return conf
}

type Consumer struct {
}

func NewConsumer() (*Consumer, error) {
	kafkaVersion, err := sarama.ParseKafkaVersion("2.1.1a")
	if err != nil {
		return nil, errors.Wrap(err, "error when parse kafka version")
	}
	_ = kafkaVersion
	consumer := &Consumer{}
	return consumer, nil
}

func (c Consumer) Read(timeout time.Duration) {

}
