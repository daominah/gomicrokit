package main

import (
	"net/url"
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/socketio"
	"github.com/daominah/gomicrokit/socketio/transport"
)

type Channel struct {
	Channel string `json:"channel"`
}

type Message struct {
	Id      int    `json:"id"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func sendJoin(c *socketio.Client) {
	log.Println("Acking /join")
	result, err := c.Ack("/join", Channel{"main"}, time.Second*5)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Ack result to /join: ", result)
	}
}

func main() {
	url0 := "ws://localhost/socket.io/"
	query := url.Values{}
	query.Add("EIO", "3")
	query.Add("transport", "websocket")
	query.Add("__sails_io_sdk_version", "1.2.1")
	url0 += "?" + query.Encode()
	log.Info(url0)

	c, err := socketio.Dial(url0, transport.GetDefaultWebsocketTransport())
	log.Infof("dial err: %#v", err)
	if err != nil {
		panic("")
	}
	defer c.Close()

	err = c.On("/message", func(h *socketio.Channel, args Message) {
		log.Println("--- Got chat message: ", args)
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(socketio.OnDisconnection, func(h *socketio.Channel) {
		log.Fatal("Disconnected")
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(socketio.OnConnection, func(h *socketio.Channel) {
		log.Println("Connected")
	})
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	go sendJoin(c)
	go sendJoin(c)
	go sendJoin(c)
	go sendJoin(c)
	go sendJoin(c)

	time.Sleep(60 * time.Second)
	c.Close()

	log.Println(" [x] Complete")
}
