package main

import (
	"time"

	"github.com/daominah/gomicrokit/kafka"
	"github.com/daominah/gomicrokit/log"
)

func main() {
	conf := kafka.ConsumerConfig{
		BootstrapServers: "127.0.0.1:9092",
		Topics:           "topic01",
		GroupId:          "group0",
		//GroupId:          fmt.Sprintf("group%v", time.Now().UnixNano()),
		Offset: kafka.OffsetEarliest}
	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i > -1; i++ {
		//log.Debugf("i", i)
		msg, err := consumer.ReadMessage(500 * time.Millisecond)
		if err != nil {
			//log.Infof("error in consumer read: %v", err)
		}
		_ = msg
		//consumer.Close()
	}
	select {}
}
