package main

import (
	"time"

	"github.com/daominah/gomicrokit/kafka"
	"github.com/daominah/gomicrokit/log"
)

func main() {
	conf := kafka.ConsumerConfig{
		//BootstrapServers: "172.18.0.201:9092,172.18.0.202:9092,172.18.0.203:9092",
		BootstrapServers:  "10.100.50.100:9092,10.100.50.101:9092,10.100.50.102:9092",
		Topics:           "topic16",
		GroupId:          "group0",
		//GroupId:          fmt.Sprintf("group%v", time.Now().UnixNano()),
		Offset: kafka.OffsetLatest}
	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i > -1; i++ {
		//log.Debugf("about to call ReadMessage %v", i)
		msg, err := consumer.ReadMessage(1000 * time.Millisecond)
		if err != nil {
			log.Infof("error in consumer read: %v", err)
		}
		_ = msg
		//consumer.Close()
	}
	select {}
}
