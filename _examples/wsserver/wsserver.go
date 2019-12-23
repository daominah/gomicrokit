package main

import (
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/websocket"
)

func main() {
	websocket.LOG = true
	websocket.SetWebsocketConfig(
		60*time.Second, 60*time.Second, 25*time.Second, 65536)

	server := websocket.NewServer(":8001", "/ws",
		nil, nil)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
