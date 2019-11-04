package websocket

import (
	"fmt"
	"time"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/maths"
	goraws "github.com/gorilla/websocket"
)

// whether to log every ws message
var Log = true
var wscfg = wsConfig{
	WriteWait:         10 * time.Second,
	PongWait:          60 * time.Second,
	PingPeriod:        54 * time.Second,
	LimitMessageBytes: 65536,
}

// package scope config, should be set before create any connection
type wsConfig struct {
	// Time allowed to write a message to the peer
	WriteWait time.Duration
	// Time allowed to read the next pong message from the peer
	PongWait time.Duration
	// Send pings to peer with this period. Must be less than pongWait
	PingPeriod time.Duration
	// Maximum message size allowed from peer,
	// limit exceeded cause the conn to close
	LimitMessageBytes int64
}

func SetWebsocketConfig(writeWait time.Duration, pongWait time.Duration,
	limitMessageBytes int64) {
	wscfg.WriteWait = writeWait
	wscfg.PongWait = pongWait
	wscfg.PingPeriod = pongWait * 9 / 10
	wscfg.LimitMessageBytes = limitMessageBytes
}

type ConnectionId string

type OnReadHandler interface {
	// Handle will be called in goroutine when conn received a msg from remote
	Handle(cid ConnectionId, msg string)
}

// emptyHandler does nothing
type emptyHandler struct{}

func (h *emptyHandler) Handle(cid ConnectionId, msg string) {}

// Connection wrap a gorrila_websocket_Conn
// Should be created by calling func NewConnection.
type Connection struct {
	conn *goraws.Conn
	// Handle will be called in goroutine when received a msg from remote
	OnReadHandler OnReadHandler
	// any unique string
	Id       ConnectionId
	createAt time.Time
	// using by writePump to ensure one concurrent writer.
	writeChan chan []byte
	// receive notification when ReadPump and WritePump ended
	disconnectChan chan bool
	IsDisconnected bool
}

func genConnId(goraConn *goraws.Conn) ConnectionId {
	if goraConn == nil {
		return ConnectionId(fmt.Sprintf("ws%v", maths.GenUUID()))
	}
	return ConnectionId(fmt.Sprintf("[ws%v|%v|%v]",
		goraConn.LocalAddr(), goraConn.RemoteAddr(), maths.GenUUID()[:4]))
}

// Wrap a connected gorilla websocket,
// this_Write and this_WriteBytes is safe to use in many goroutines
func NewConnection(goraConn *goraws.Conn, handler OnReadHandler) *Connection {
	if handler == nil {
		handler = &emptyHandler{}
	}
	c := &Connection{
		conn:           goraConn,
		OnReadHandler:  handler,
		Id:             genConnId(goraConn),
		createAt:       time.Now(),
		writeChan:      make(chan []byte),
		disconnectChan: make(chan bool, 2),
		IsDisconnected: false,
	}
	go c.WritePump()
	go c.ReadPump()
	go c.OnDisconnect()
	return c
}

func (c *Connection) ReadPump() {
	defer func() {
		c.conn.Close()
		log.Condf(Log, "read pump of %v returned", c.Id)
		c.disconnectChan <- true
	}()
	c.conn.SetReadLimit(wscfg.LimitMessageBytes)
	c.conn.SetReadDeadline(time.Now().Add(wscfg.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(wscfg.PongWait))
		return nil
	})
	for {
		_, messageB, err := c.conn.ReadMessage()
		if err != nil {
			log.Condf(Log, "error when %v read message: %v", c.Id, err)
			break
		}
		msg := string(messageB)
		log.Condf(Log, "received a message from %v: %v", c.Id, msg)
		go c.OnReadHandler.Handle(c.Id, msg)
	}
}

// Ensure there is at most one writer to a connection by  executing all writes
// from this goroutine.
func (c *Connection) WritePump() {
	ticker := time.NewTicker(wscfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Condf(Log, "write pump of %v returned", c.Id)
		c.disconnectChan <- true
	}()
	for {
		var msgB []byte
		select {
		case msgB = <-c.writeChan:
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.TextMessage, msgB)
			if err != nil {
				log.Condf(Log, "error when write to %v: %v", c.Id, err)
				return
			}
		case <-ticker.C:
			msgB = []byte("PING")
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.PingMessage, nil)
			if err != nil {
				log.Condf(Log, "error when ping to %v: %v", c.Id, err)
				return
			}
		}
		log.Condf(Log, "wrote to %v msg: %v", c.Id, string(msgB))
	}
}

func (c *Connection) OnDisconnect() {
	<-c.disconnectChan
	c.IsDisconnected = true
	log.Condf(Log, "connection %v disconnected", c.Id)
}

func (c *Connection) WriteBytes(message []byte) {
	if c.IsDisconnected {
		return
	}
	timeout := time.After(3 * time.Second)
	select {
	case c.writeChan <- message:
	case <-timeout:
		log.Infof("timeout when send to write chan of %v", c.Id)
	}
}

func (c *Connection) Write(message string) {
	c.WriteBytes([]byte(message))
}

func (c *Connection) Close() {
	log.Condf(Log, "about to close %v", c.Id)
	c.conn.Close()
}
