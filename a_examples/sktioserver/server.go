package main

import (
	"log"
	"net/http"
	"time"

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

func main() {
	server := socketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(socketio.OnConnection, func(c *socketio.Channel) {
		log.Println("Connected")

		c.Emit("/message", Message{10, "main", "using emit"})

		c.Join("test")
		c.BroadcastTo("test", "/message", Message{10, "main", "using broadcast"})
	})
	server.On(socketio.OnDisconnection, func(c *socketio.Channel) {
		log.Println("Disconnected")
	})

	server.On("/join", func(c *socketio.Channel, channel Channel) string {
		time.Sleep(2 * time.Second)
		log.Println("Client joined to ", channel.Channel)
		return "joined to " + channel.Channel
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	log.Println("Starting server...")
	log.Panic(http.ListenAndServe(":3811", serveMux))
}
