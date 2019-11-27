package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/daominah/gomicrokit/log"
	goraws "github.com/gorilla/websocket"
)

// whether to log every ws message
var LOG = true

// wscfg is this package global config for reading and writing messages
var wscfg = wsConfig{
	WriteWait:         60 * time.Second,
	PongWait:          60 * time.Second,
	PingPeriod:        25 * time.Second,
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

// change the config of this package for reading and writing messages
func SetWebsocketConfig(writeWait time.Duration, pongWait time.Duration,
	pingPeriod time.Duration, limitMessageBytes int64) {
	wscfg.WriteWait = writeWait
	wscfg.PongWait = pongWait
	wscfg.PingPeriod = pingPeriod
	wscfg.LimitMessageBytes = limitMessageBytes
}

// ConnectionId is anything unique for each connection object
type ConnectionId string

// GenConnId generates ConnectionId by concat local and remote address,
// the ConnectionId is unique for each active connection
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

type OnReadHandler interface {
	// Handle will be called in a goroutine when conn received a msg from remote.
	// :param msgType: int, RFC 6455: TextMessage = 1, BinaryMessage = 2, ..
	Handle(cid ConnectionId, msgType int, msg []byte)
}

// emptyHandler implements OnReadHandler, this handle does nothing
type emptyHandler struct{}

func (h *emptyHandler) Handle(cid ConnectionId, msgType int, msg []byte) {}

// Connection wraps a gorrila_websocket_Conn,
// conn_WriteBytes and conn_Write is safe for concurrent calls
type Connection struct {
	conn *goraws.Conn
	// Handle will be called in goroutine when received a msg from remote
	OnReadHandler OnReadHandler
	id            ConnectionId
	// writeChan is using by this_writePump to ensure one concurrent writer
	writeChan chan []byte
	// closedChan will be closed when this connection disconnected
	closedChan <-chan struct{}
	// call notifyClosed() will close the closedChan,
	// after the first call, subsequent calls to this func do nothing
	notifyClosed context.CancelFunc
}

func dial(wsServerAddr string, skipTls bool) (*goraws.Conn, error) {
	dialer := *goraws.DefaultDialer
	if skipTls {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	goraConn, _, err := dialer.Dial(wsServerAddr, nil)
	if err == nil {
		log.Condf(LOG, "%v connected", GenConnId(goraConn))
	}
	return goraConn, err
}

// Dial creates a gorrila_websocket_Conn
func Dial(wsServerAddr string) (*goraws.Conn, error) {
	return dial(wsServerAddr, false)
}

// DialSkipTls creates a gorrila_websocket_Conn. Using in testing wss host.
// This func accepts any certificate presented by the server and any host name
// in that certificate. In this mode, TLS is susceptible to man-in-the-middle
// attacks
func DialSkipTls(wsServerAddr string) (*goraws.Conn, error) {
	return dial(wsServerAddr, true)
}

// :param goraConn: a gorrila_websocket_Conn, can be created by functions Dial
// or DialSkipTls of this packages.
// :param onRead: a handler, its method Handler will be called in a goroutine
// for each received a msg from remote.
func NewConnection(goraConn *goraws.Conn, onRead OnReadHandler) *Connection {
	if onRead == nil {
		onRead = &emptyHandler{}
	}
	ctx, calcel := context.WithCancel(context.Background())
	c := &Connection{
		conn:          goraConn,
		OnReadHandler: onRead,
		id:            GenConnId(goraConn),
		writeChan:     make(chan []byte),
		closedChan:    ctx.Done(),
		notifyClosed:  calcel,
	}
	go c.writePump()
	go c.readPump()
	go func() {
		<-c.closedChan
		log.Condf(LOG, "%v disconnected", c.id)
	}()
	return c
}

func (c *Connection) readPump() {
	defer func() {
		log.Condf(LOG, "read pump of %v returned", c.id)
		c.Close()
	}()
	c.conn.SetReadLimit(wscfg.LimitMessageBytes)
	c.conn.SetReadDeadline(time.Now().Add(wscfg.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(wscfg.PongWait))
		return nil
	})
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Condf(LOG, "error when %v read message: %v", c.id, err)
			break
		}
		log.Condf(LOG, "received a message from %v: %s", c.id, msg)
		go c.OnReadHandler.Handle(c.id, msgType, msg)
	}
}

// writePump ensures there is at most one write to a connection at a moment
func (c *Connection) writePump() {
	ticker := time.NewTicker(wscfg.PingPeriod)
	defer func() {
		log.Condf(LOG, "write pump of %v returned", c.id)
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case msgB := <-c.writeChan:
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.TextMessage, msgB)
			if err != nil {
				log.Condf(LOG, "error when write to %v: %v", c.id, err)
				return
			}
			log.Condf(LOG, "wrote to %v msg: %v", c.id, string(msgB))
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(wscfg.WriteWait))
			err := c.conn.WriteMessage(goraws.PingMessage, nil)
			if err != nil {
				log.Condf(LOG, "error when ping to %v: %v", c.id, err)
				return
			}
		case <-c.closedChan:
			return
		}
	}
}

func (c *Connection) WriteBytes(message []byte) {
	timeout := time.After(3 * time.Second)
	select {
	case c.writeChan <- message:
	case <-timeout:
		log.Condf(LOG, "timeout when send to write chan of %v", c.id)
	case <-c.closedChan:
		log.Condf(LOG, "write to closed connection %v", c.id)
	}
}

func (c *Connection) Write(message string) {
	c.WriteBytes([]byte(message))
}

func (c *Connection) Close() {
	c.conn.Close()
	c.notifyClosed()
}
