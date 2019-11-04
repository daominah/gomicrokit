package main

import (
	"github.com/daominah/gomicrokit/kafka"
	"github.com/daominah/gomicrokit/log"
	"fmt"
	"time"
)

func main() {
	conf := kafka.ProducerConfig{
		BrokersList:  "127.0.0.1:9092",
		DefaultTopic: "topic01",
	}
	producer, err := kafka.NewProducer(conf)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i > -1; i++ {
		time.Sleep(1000 * time.Millisecond)
		err := producer.SendMessage(fmt.Sprintf("pussy %v", i))
		if err != nil {
			log.Info(err)
		}
	}
}
