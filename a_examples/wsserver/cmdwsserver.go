package main

import (
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/websocket"
)

func main() {
	websocket.LOG = true
	customConfig := &websocket.Config{
		WriteWait:         20 * time.Second,
		PongWait:          20 * time.Second,
		PingPeriod:        8 * time.Second,
		LimitMessageBytes: 65536,
	}

	server := websocket.NewServer(":8001", "/", customConfig)
	server.Handler = websocket.Ignorer{}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
