package kafka

import (
	"context"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/daominah/gomicrokit/log"
	"github.com/pkg/errors"
)

type Offset int64

const (
	OffsetEarliest = Offset(sarama.OffsetOldest)
	OffsetLatest   = Offset(sarama.OffsetNewest)
)

var (
	ErrReadMsgTimeout     = errors.New("read message timeout")
	ErrReadClosedConsumer = errors.New("read message on closed consumer")
)

// ConsumerConfig should be created by NewConsumerConfig (for default values)
type ConsumerConfig struct {
	// comma separated list: broker1:9092,broker2:9092,broker3:9092
	BootstrapServers string
	// comma separated list topics to subscribe: topic0,topic1,topic2
	Topics  string
	GroupId string
	Offset  Offset
}

// represent consumed message from kafka
type Message struct {
	Value     string
	Offset    int64
	Topic     string
	Partition int32
	Key       string
	Timestamp time.Time
}

type Consumer struct {
	client  sarama.ConsumerGroup
	handler *ConsumerGroupHandlerImpl
	// call this func to stop the connecting loop in the constructor
	cancelFunc context.CancelFunc
	closed     bool
	// help to only log the first time try to reconnect to kafka,
	// reset by a successfully connect
	isTryingReconnect bool
}

func NewConsumer(conf ConsumerConfig) (*Consumer, error) {
	log.Infof("creating a consumer from %#v", conf)
	// construct sarama config
	kafkaVersion, err := sarama.ParseKafkaVersion("1.1.1")
	if err != nil {
		return nil, errors.Wrap(err, "error when parse kafka version")
	}
	samConf := sarama.NewConfig()
	samConf.Version = kafkaVersion
	samConf.Consumer.Offsets.Initial = int64(conf.Offset)

	// connect to kafka
	c := &Consumer{}
	brokers := strings.Split(conf.BootstrapServers, ",")
	c.client, err = sarama.NewConsumerGroup(brokers, conf.GroupId, samConf)
	if err != nil {
		return nil, errors.Wrap(err, "err when create consumer client")
	}
	log.Infof("connected to kafka cluster %v", conf.BootstrapServers)

	c.handler = &ConsumerGroupHandlerImpl{
		consumer:    c,
		readyChan:   make(chan bool),
		readMsgChan: make(chan *ReadMsgRequest),
	}
	var ctx context.Context
	ctx, c.cancelFunc = context.WithCancel(context.Background())
	topics := strings.Split(conf.Topics, ",")
	go func() {
		// below loop create new session if kafka server rebalance comsumers
		for {
			if !c.isTryingReconnect {
				log.Infof("joining consumer group")
			}
			err := c.client.Consume(ctx, topics, c.handler)
			// session ended
			if err != nil {
				if !c.isTryingReconnect {
					log.Infof("error when kafka consume: %v", err)
				}
				c.isTryingReconnect = true
				// wait a second then try to reconnect
				time.Sleep(time.Second)
				continue
			}
			if ctx.Err() != nil {
				log.Infof("ctx cancelled when kafka consume: %v", err)
				return
			}
			log.Infof("session ended normally (probably because kafka " +
				"server rebalance cycle is initiated)")
			// avoid panic because close readyChan twice in chl.Setup
			c.handler.readyChan = make(chan bool)
		}
	}()
	// wait client to join consumer group
	<-c.handler.readyChan
	return c, nil
}

func (c *Consumer) Close() {
	c.closed = true
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	if c.client != nil {
		log.Debugf("consumer_Close cp1")
		err := c.client.Close()
		// TODO: hanging here
		log.Debugf("error when close client: %v", err)
	}
	log.Debugf("consumer_Close cp2")

}

type ReadMsgRequest struct {
	ctx          context.Context
	responseChan chan *Message
}

type ConsumerGroupHandlerImpl struct {
	consumer *Consumer
	// close this channel to notify client joined consumer group successfully
	readyChan chan bool
	// send to this channel to read a kafka message
	readMsgChan chan *ReadMsgRequest
}

func (h *ConsumerGroupHandlerImpl) Setup(s sarama.ConsumerGroupSession) error {
	log.Infof("joined consumer group, assigned partitions %#v", s.Claims())
	h.consumer.isTryingReconnect = false
	close(h.readyChan)
	return nil
}

func (h *ConsumerGroupHandlerImpl) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// each assigned partition will run this func in a goroutine
func (h *ConsumerGroupHandlerImpl) ConsumeClaim(
	session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for i := 0; i > -1; i++ {
		readRequest := <-h.readMsgChan
		select {
		case samMsg, opening := <-claim.Messages():
			if !opening {
				log.Infof("handler func ConsumeClaim returned")
				return nil
			}
			if samMsg != nil {
				msg := &Message{Value: string(samMsg.Value), Offset: samMsg.Offset,
					Topic: samMsg.Topic, Partition: samMsg.Partition,
					Key: string(samMsg.Key), Timestamp: samMsg.Timestamp}
				log.Infof("received a message from topic %v:%v:%v: %v",
					msg.Topic, msg.Partition, msg.Offset, msg.Value)
				select {
				case readRequest.responseChan <- msg:
					session.MarkMessage(samMsg, "")
				case <-readRequest.ctx.Done():
					log.Infof("timed out when reply to read message")
				}
			} else {
				log.Debugf("huh?")
			}
		case <-readRequest.ctx.Done():
			// read message request timed out
			continue
		}
	}
	return nil
}

// timeout: maximum time to block waiting for message,
//     set timeout = -1 to wait forever
func (c Consumer) ReadMessage(timeout time.Duration) (*Message, error) {
	if c.closed {
		return nil, ErrReadClosedConsumer
	}
	if timeout < 0 {
		timeout = 24 * time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	request := &ReadMsgRequest{ctx: ctx, responseChan: make(chan *Message)}
	select {
	case c.handler.readMsgChan <- request:
		select {
		case msg := <-request.responseChan:
			return msg, nil
		case <-ctx.Done():
			return nil, ErrReadMsgTimeout
		}
	case <-ctx.Done():
		// because client disconnected to kafka so ConsumeClaim is not running
		// or timeout duration is too short
		//log.Debugf("time out when send to c_handler_readMsgChan")
		return nil, ErrReadMsgTimeout
	}
}
