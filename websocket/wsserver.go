package websocket

import (
	"net/http"
	"sync"

	"github.com/daominah/gomicrokit/httpsvr"
	"github.com/daominah/gomicrokit/log"
	goraws "github.com/gorilla/websocket"
)

// ServerHandler defines events a websocket server must support
type ServerHandler interface {
	OnMessageHandler
	// OnOpen will be called after a client connected
	OnOpen(cid ConnectionId, initHttpReq *http.Request)
	// OnClose will be called after a connection disconnected,
	// OnClose acts as both onerror and onclose in W3C standards
	OnClose(cid ConnectionId)
}

// Server must be initialized by calling NewServer
type Server struct {
	// Handler implements OnOpen, OnMessage, OnClose functions,
	// the Handler must be assigned before call this_ListenAndServe
	Handler ServerHandler
	// in case you want to use the port as a http server too
	HttpRouter    *httpsvr.Server
	wsConfig      Config
	listeningPort string // begin with char ":"
	wsPath        string // begin with char "/"
	connections   map[ConnectionId]*Connection
	mutex         *sync.Mutex
}

// NewServer initializes a Server,
// :param wsConfig: can be nil
func NewServer(listeningPort string, wsPath string, wsConfig *Config) *Server {
	if wsConfig == nil {
		wsConfig = &DefaultConfig
	}
	s := &Server{
		Handler:       &Ignorer{},
		HttpRouter:    httpsvr.NewServer(),
		wsConfig:      *wsConfig,
		listeningPort: listeningPort,
		wsPath:        wsPath,
		connections:   make(map[ConnectionId]*Connection),
		mutex:         &sync.Mutex{},
	}
	s.HttpRouter.AddHandler("GET", wsPath, s.handleUpgradeWs())
	return s
}

// ListenAndServe listens on the listeningPort
func (s Server) ListenAndServe() error {
	log.Infof(`starting websocket server on "ws://host%v%v`,
		s.listeningPort, s.wsPath)
	return s.HttpRouter.ListenAndServe(s.listeningPort)
}

// WriteAll sends a TextMessage to all clients
func (s Server) WriteAll(message string) {
	s.mutex.Lock()
	for _, conn := range s.connections {
		cloned := conn
		if cloned != nil {
			go cloned.Write(message)
		}
	}
	s.mutex.Unlock()
}

// WriteBytesAll sends a BinaryMessage to all clients
func (s Server) WriteBytesAll(message []byte) {
	s.mutex.Lock()
	for _, conn := range s.connections {
		cloned := conn
		if cloned != nil {
			go cloned.WriteBytes(message)
		}
	}
	s.mutex.Unlock()
}

// Write sends a TextMessage to the given connection
func (s Server) Write(connId ConnectionId, message string) {
	s.mutex.Lock()
	conn := s.connections[connId]
	s.mutex.Unlock()
	if conn != nil {
		conn.Write(message)
	}
}

// WriteBytes sends a BinaryMessage to the given connection
func (s Server) WriteBytes(connId ConnectionId, message []byte) {
	s.mutex.Lock()
	conn := s.connections[connId]
	s.mutex.Unlock()
	if conn != nil {
		conn.WriteBytes(message)
	}
}

func (s *Server) handleUpgradeWs() http.HandlerFunc {
	upgrader := goraws.Upgrader{
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	return func(w http.ResponseWriter, r *http.Request) {
		goraConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Infof("cannot Upgrade for %v: %v", r.RemoteAddr, err)
			return
		}
		conn := wrapConn(goraConn, s.Handler)
		conn.Config = s.wsConfig
		s.mutex.Lock()
		s.connections[conn.Id] = conn
		s.mutex.Unlock()
		go s.Handler.OnOpen(conn.Id, r)
		log.Condf(LOG, "nWSConnections: %v", s.GetNumberConnections())
		go func() {
			<-conn.ClosedCtxDone
			s.mutex.Lock()
			delete(s.connections, conn.Id)
			s.mutex.Unlock()
			go s.Handler.OnClose(conn.Id)
			log.Condf(LOG, "nWSConnections: %v", s.GetNumberConnections())
		}()
	}
}

// GetConnection can return nil
func (s Server) GetConnection(connId ConnectionId) *Connection {
	s.mutex.Lock()
	conn := s.connections[connId]
	s.mutex.Unlock()
	return conn
}

// GetNumberConnections returns number of connecting clients
func (s Server) GetNumberConnections() int {
	s.mutex.Lock()
	ret := len(s.connections)
	s.mutex.Unlock()
	return ret
}
