package main

import (
	"fmt"
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/websocket"
)

func main() {
	websocket.Log = true
	websocket.SetWebsocketConfig(2*time.Second, 5*time.Second, 65536)

	for k := 0; k < 3; k++ {
		goraConn, err := websocket.Dial("ws://127.0.0.1:8000/")
		if err != nil {
			log.Infof("error when ws dial: %v", err)
			continue
		}
		conn := websocket.NewConnection(goraConn, nil, nil)
		for i := 0; i < 3; i++ {
			go func(m int) {
				conn.Write(fmt.Sprintf("%v", m))
			}(10*k + i)
		}
	}
	log.Infof("check point bottom")
	select {}
}
