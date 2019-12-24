package main

import (
	"fmt"
	"time"

	"github.com/daominah/gomicrokit/kafka"
	"github.com/daominah/gomicrokit/log"
)

func main() {
	conf := kafka.ProducerConfig{
		BrokersList:  "127.0.0.1:9092",
		DefaultTopic: "topic11",
	}
	producer, err := kafka.NewProducer(conf)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i > -1; i++ {
		_ = time.Millisecond
		//time.Sleep(2 * time.Millisecond)
		err := producer.SendMessage(fmt.Sprintf("pussy %v", i))
		//err := producer.SendExplicitMessage(conf.DefaultTopic, fmt.Sprintf("pussy %v", i), "key0")
		if err != nil {
			log.Info(err)
		}
		//break
	}
	select {}
}
