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
	for i := 0; i > -1; i++ {
		time.Sleep(500 * time.Millisecond)
		go func(i int) {
			_ = time.Millisecond
			err := producer.SendMessage(fmt.Sprintf("pussy %v %v", i, time.Now().Format(time.RFC3339)))
			//err := producer.SendExplicitMessage(conf.DefaultTopic, fmt.Sprintf("pussy %v", i), "key0")
			if err != nil {
				log.Info(err)
			}
		}(i)
		//break
	}
	log.Infof("the loop ended")
	select {}
}
