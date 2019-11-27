package main

import (
	"fmt"
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/websocket"
)

func main() {
	websocket.LOG = true
	websocket.SetWebsocketConfig(
		60*time.Second, 60*time.Second, 25*time.Second, 65536)

	// k is number of connections to create
	for k := 0; k < 2; k++ {
		go func(k int) {
			goraConn, err := websocket.Dial("ws://127.0.0.1:8000/")
			if err != nil {
				log.Infof("error when ws dial: %v", err)
				return
			}
			conn := websocket.NewConnection(goraConn, nil)
			for i := 0; i < 4; i++ {
				time.Sleep(1 * time.Second)
				if i == 2 {
					conn.Close()
				} else {
					go func(i10k int) {
						if k%2 == 0 {
							conn.Write(fmt.Sprintf("%v", i10k))
						} else {
							conn.WriteBytes([]byte(fmt.Sprintf("%v", i10k)))
						}
					}(10*k + i)
				}
			}
		}(k)
	}
	log.Infof("check point bottom")
	select {}
}
