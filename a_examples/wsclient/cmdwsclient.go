package main

import (
	"fmt"
	"time"

	"sync"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/websocket"
)

func main() {
	conn, err := websocket.NewConnection(
		"ws://127.0.0.1:8001/", nil, true)
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		if i == 2 {
			conn.Close()
		} else {
			wg.Add(1)
			go func(i int) {
				defer wg.Add(-1)
				log.Infof("loop %v", i)
				if i%2 == 0 {
					conn.Write(fmt.Sprintf("%v", i))
				} else {
					conn.WriteBytes([]byte(fmt.Sprintf("%v", i)))
				}
			}(i)
		}
	}
	wg.Wait()
	log.Infof("main returned")
}
