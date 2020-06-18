package main

import (
	"fmt"
	"time"

	"github.com/daominah/gomicrokit/kafka"
	"github.com/daominah/gomicrokit/log"
)

func main() {
	conf := kafka.ProducerConfig{
		//BrokersList:  "172.18.0.201:9092,172.18.0.202:9092,172.18.0.203:9092",
		BrokersList:  "10.100.50.100:9092,10.100.50.101:9092,10.100.50.102:9092",
		DefaultTopic: "topic16",
	}
	producer, err := kafka.NewProducer(conf)
	if err != nil {
		log.Fatal(err)
	}
	msg := fmt.Sprintf("pussy %v", time.Now().Format(time.RFC3339))
	err = producer.SendMessage(msg)
	if err != nil {
		log.Info(err)
	}
	select {}
}
