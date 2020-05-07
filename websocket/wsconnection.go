// Package websocket is an easy-to-use websocket client and server.
// This package tries to follow W3C standards [w3c.github.io/websockets/]
package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daominah/gomicrokit/log"
	goraws "github.com/gorilla/websocket"
)

// LOG determines whether to log every ws message
var LOG = true

// Config defines heart beats duration, limits for sending message,
// usually the DefaultConfig is good enough
type Config struct {
	// Time allowed to write a message to the peer
	WriteWait time.Duration
	// Time allowed to read the next pong message from the peer
	PongWait time.Duration
	// Send pings to peer with this period, must be less than PongWait
	PingPeriod time.Duration
	// Maximum message size allowed from peer, excess causes the conn to close
	LimitMessageBytes int64
}

// DefaultConfig defines heart beats duration, limits for sending message
var DefaultConfig = Config{
	WriteWait:         60 * time.Second,
	PongWait:          60 * time.Second,
	PingPeriod:        25 * time.Second,
	LimitMessageBytes: 16384,
}

// OnMessageHandler wraps func OnMessage
type OnMessageHandler interface {
	// onMessage is a function that will be called in a goroutine when
	// a message is received from remote.
	// :param msgType: int, RFC 6455: TextMessage = 1, BinaryMessage = 2, ..
	// :param cid: only for server codes, client codes can ignore this param
	OnMessage(msg []byte, msgType int, cid ConnectionId)
}

// Connection wraps a gorrila_websocket_Conn,
// Connection must be init by calling NewConnection,
// conn_WriteBytes and conn_Write is safe for concurrent calls
type Connection struct {
	Config Config       // can be changed after the connection init
	Id     ConnectionId // auto generated when init, should not be changed
	conn   *goraws.Conn
	// onMessage will be called in goroutine when received a msg from remote,
	// default value is Ignorer
	onMessage OnMessageHandler
	// writeChan is using by this_writePump to ensure one concurrent writer
	writeChan chan *wsMessage
	// receiving on this channel to know when this connection closed
	ClosedCtxDone <-chan struct{}
	// call closedCxl() will close the ClosedCtxDone channel
	closedCxl context.CancelFunc
}

// NewConnection initializes a Connection,
// :param url: example: ws://127.0.0.1:8001/,
// :param onMessage: can be nil,
// :param isSkipTLS: should only be true in testing environment, accepts any
// certificate presented by the server and any host name in that certificate,
// in this mode, TLS is susceptible to man-in-the-middle attacks,
func NewConnection(url string, onMessage OnMessageHandler, isSkipTLS bool) (
	*Connection, error) {
	dialer := *goraws.DefaultDialer
	if isSkipTLS {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	gorillaConn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return wrapConn(gorillaConn, onMessage), nil
}

func wrapConn(gorillaConn *goraws.Conn, onMessage OnMessageHandler) *Connection {
	if onMessage == nil {
		onMessage = &Ignorer{}
	}
	ctx, cxl := context.WithCancel(context.Background())
	c := &Connection{
		Config:        DefaultConfig,
		Id:            GenConnId(gorillaConn),
		conn:          gorillaConn,
		onMessage:     onMessage,
		writeChan:     make(chan *wsMessage),
		ClosedCtxDone: ctx.Done(),
		closedCxl:     cxl,
	}
	go c.writePump()
	go c.readPump()
	log.Condf(LOG, "connected %v", c.Id)
	go func() {
		<-c.ClosedCtxDone
		log.Condf(LOG, "disconnected %v", c.Id)
	}()
	return c
}

// Send sends a TextMessage to remote
func (c Connection) Send(message string) { c.Write(message) }

// Close closes the connection without sending or waiting for a close message
func (c *Connection) Close() {
	c.conn.Close()
	c.closedCxl()
}

// WriteBytes sends a BinaryMessage to remote
func (c Connection) WriteBytes(message []byte) {
	c.writeBytes([]byte(message), true)
}

// Write sends a TextMessage to remote
func (c Connection) Write(message string) {
	c.writeBytes([]byte(message), false)
}

// CheckIsClosed returns true if the connection is disconnected
func (c Connection) CheckIsClosed() bool {
	select {
	case <-c.ClosedCtxDone:
		return true
	default:
		return false
	}
}

func (c Connection) writeBytes(message []byte, isBinMsg bool) {
	timeout := time.After(c.Config.WriteWait)
	select {
	case c.writeChan <- &wsMessage{data: message, isBinaryMessage: isBinMsg}:
		// pass
	case <-timeout:
		log.Condf(LOG, "timed out when send to writeChan of %v", c.Id)
	case <-c.ClosedCtxDone:
		log.Condf(LOG, "cannot write on closed connection %v", c.Id)
	}
}

func (c *Connection) readPump() {
	c.conn.SetReadLimit(c.Config.LimitMessageBytes)
	c.conn.SetReadDeadline(time.Now().Add(c.Config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.Config.PongWait))
		return nil
	})
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Condf(LOG, "cannot read message from %v: %v", c.Id, err)
			break
		}
		log.Condf(LOG, "received a message from %v: %s", c.Id, msg)
		go c.onMessage.OnMessage(msg, msgType, c.Id)
	}
	c.closedCxl()
	log.Condf(LOG, "readPump of %v returned", c.Id)
}

// writePump ensures there is at most one write to a connection at a moment
func (c *Connection) writePump() {
	ticker := time.NewTicker(c.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.closedCxl()
		log.Condf(LOG, "write pump of %v returned", c.Id)
	}()
	for {
		select {
		case wsMsg := <-c.writeChan:
			c.conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			var err error
			if wsMsg.isBinaryMessage {
				err = c.conn.WriteMessage(goraws.BinaryMessage, wsMsg.data)
			} else {
				err = c.conn.WriteMessage(goraws.TextMessage, wsMsg.data)
			}
			if err != nil {
				log.Condf(LOG, "error when write to %v: %v", c.Id, err)
				return
			}
			log.Condf(LOG, "wrote to %v msg: %s", c.Id, wsMsg.data)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			err := c.conn.WriteMessage(goraws.PingMessage, nil)
			if err != nil {
				log.Condf(LOG, "error when ping to %v: %v", c.Id, err)
				return
			}
		case <-c.ClosedCtxDone:
			return
		}
	}
}

// ConnectionId should be unique for each Connection,
// ConnectionId is auto generated in func NewConnection
type ConnectionId string

// GenConnId generates ConnectionId by concat local and remote address,
// so the ConnectionId is unique for each active connection.
// UUID is better for unique but is much worse for self-describing
func GenConnId(goraConn *goraws.Conn) ConnectionId {
	if goraConn == nil {
		return ConnectionId(fmt.Sprintf("[ws|nil]"))
	}
	localAddr := goraConn.LocalAddr().String()
	colon := strings.Index(localAddr, ":")
	if colon != -1 {
		// localAddr is only "port" instead of "ip:port"
		localAddr = localAddr[colon:]
	}
	return ConnectionId(fmt.Sprintf("[ws|%v|%v]",
		localAddr, goraConn.RemoteAddr()))
}

type wsMessage struct {
	data            []byte
	isBinaryMessage bool
}

// Ignorer implements OnMessageHandler, this handler does nothing,
// Ignorer implements ServerHandler too
type Ignorer struct{}

func (h Ignorer) OnMessage(msg []byte, msgType int, cid ConnectionId) {}
func (h Ignorer) OnOpen(cid ConnectionId, initHttpReq *http.Request)  {}
func (h Ignorer) OnClose(cid ConnectionId)                            {}
