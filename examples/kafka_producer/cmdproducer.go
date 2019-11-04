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
		DefaultTopic: "topic05",
	}
	producer, err := kafka.NewProducer(conf)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i > -1; i++ {
		time.Sleep(450 * time.Millisecond)
		err := producer.SendMessage(fmt.Sprintf("pussy %v", i))
		//err := producer.SendExplicitMessage(conf.DefaultTopic, fmt.Sprintf("pussy %v", i), "key0")
		if err != nil {
			log.Info(err)
		}
	}
}
