package websocket

import (
	"context"
	"fmt"
	"strings"
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
	id       ConnectionId
	createAt time.Time
	// using by writePump to ensure one concurrent writer.
	writeChan chan []byte
	// closedChan will be closed by this_notifyClosed when the conn disconnected
	closedChan   <-chan struct{}
	notifyClosed context.CancelFunc
}

func genConnId(goraConn *goraws.Conn) ConnectionId {
	if goraConn == nil {
		return ConnectionId(fmt.Sprintf("[ws|%v]", maths.GenUUID()))
	}
	localAddr := goraConn.LocalAddr().String()
	colon := strings.Index(localAddr, ":")
	if colon != -1 {
		localAddr = localAddr[colon:]
	}
	return ConnectionId(fmt.Sprintf("[ws|%v|%v|%v]",
		localAddr, goraConn.RemoteAddr(), maths.GenUUID()[:4]))
}

func Dial(wsServerAddr string) (*goraws.Conn, error) {
	goraConn, _, err := goraws.DefaultDialer.Dial(wsServerAddr, nil)
	return goraConn, err
}

// Wrap a connected gorilla websocket,
// this_Write and this_WriteBytes is safe to use in many goroutines
func NewConnection(goraConn *goraws.Conn, onRead OnReadHandler,
	onDisconnect func(*Connection)) *Connection {
	if onRead == nil {
		onRead = &emptyHandler{}
	}
	ctx, calcel := context.WithCancel(context.Background())
	c := &Connection{
		conn:          goraConn,
		OnReadHandler: onRead,
		id:            genConnId(goraConn),
		createAt:      time.Now(),
		writeChan:     make(chan []byte),
		closedChan:    ctx.Done(),
		notifyClosed:  calcel,
	}
	go c.writePump()
	go c.readPump()
	go c.onDisconnect(onDisconnect)
	log.Condf(Log, "connection %v started", c.id)
	return c
}

func (c *Connection) readPump() {
	defer func() {
		c.conn.Close()
		c.notifyClosed()
		log.Condf(Log, "read pump of %v returned", c.id)
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
			log.Condf(Log, "error when %v read message: %v", c.id, err)
			break
		}
		msg := string(messageB)
		log.Condf(Log, "received a message from %v: %v", c.id, msg)
		go c.OnReadHandler.Handle(c.id, msg)
	}
}

// Ensure there is at most one writer to a connection by  executing all writes
// from this goroutine.
func (c *Connection) writePump() {
	ticker := time.NewTicker(wscfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.notifyClosed()
		log.Condf(Log, "write pump of %v returned", c.id)
	}()
	for {
		select {
		case msgB := <-c.writeChan:
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.TextMessage, msgB)
			if err != nil {
				log.Condf(Log, "error when write to %v: %v", c.id, err)
				return
			}
			log.Condf(Log, "wrote to %v msg: %v", c.id, string(msgB))
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.PingMessage, nil)
			if err != nil {
				log.Condf(Log, "error when ping to %v: %v", c.id, err)
				return
			}
		}

	}
}

func (c *Connection) onDisconnect(callback func(*Connection)) {
	<-c.closedChan
	log.Condf(Log, "connection %v disconnected", c.id)
	if callback == nil {
		return
	}
	callback(c)
}

func (c *Connection) WriteBytes(message []byte) {
	timeout := time.After(3 * time.Second)
	select {
	case c.writeChan <- message:
	case <-timeout:
		log.Infof("timeout when send to write chan of %v", c.id)
	}
}

func (c *Connection) Write(message string) {
	c.WriteBytes([]byte(message))
}

func (c *Connection) Close() {
	log.Condf(Log, "about to close %v", c.id)
	c.conn.Close()
	c.notifyClosed()
}
