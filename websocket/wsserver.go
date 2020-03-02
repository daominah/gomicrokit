package websocket

import (
	"net/http"
	"sync"

	"github.com/daominah/gomicrokit/httpsvr"
	"github.com/daominah/gomicrokit/log"
	goraws "github.com/gorilla/websocket"
)

// Server must be inited by NewServer
type Server struct {
	listeningPort string
	wsPath        string
	HttpRouter    *httpsvr.Server
	Handler       ServerHandler
	Connections   map[ConnectionId]*Connection
	Mutex         sync.Mutex
}

// NewServer returns a Server with inited map connections and router
func NewServer(listeningPort string, wsPath string) *Server {
	s := &Server{
		listeningPort: listeningPort,
		wsPath:        wsPath,
		HttpRouter:    httpsvr.NewServer(),
		Connections:   make(map[ConnectionId]*Connection),
		Handler:       &EmptyHandler{},
	}
	s.HttpRouter.AddHandler("GET", wsPath, s.handleUpgradeWs())
	return s
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
			log.Infof("error when upgrader_Upgrade for %v: %v",
				r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.Mutex.Lock()
		conn := NewConnection(goraConn, s.Handler)
		s.Connections[conn.id] = conn
		s.Mutex.Unlock()
		log.Condf(LOG, "%v connected", conn.id)
		go s.Handler.OnOpen(conn.id, r)
		go func() {
			<-conn.ClosedChan
			s.Mutex.Lock()
			delete(s.Connections, conn.id)
			s.Mutex.Unlock()
			log.Condf(LOG, "%v disconnected", conn.id)
			go s.Handler.OnClose(conn.id)
		}()
	}
}

// ListenAndServe listens on the server's listeningPort
func (s *Server) ListenAndServe() error {
	log.Infof(`starting websocket server on "ws://host%v%v`,
		s.listeningPort, s.wsPath)
	return s.HttpRouter.ListenAndServe(s.listeningPort)
}

func (s *Server) Write(connId ConnectionId, message string) {
	s.Mutex.Lock()
	conn := s.Connections[connId]
	s.Mutex.Unlock()
	if conn != nil {
		conn.Write(message)
	}
}

// WriteBytes sends a BinaryMessage to the given connection
func (s *Server) WriteBytes(connId ConnectionId, message []byte) {
	s.Mutex.Lock()
	conn := s.Connections[connId]
	s.Mutex.Unlock()
	if conn != nil {
		conn.WriteBytes(message)
	}
}

// WriteAll sends a TextMessage to all connections
func (s *Server) WriteAll(message string) {
	s.Mutex.Lock()
	for _, conn := range s.Connections {
		cloned := conn
		if cloned != nil {
			go cloned.Write(message)
		}
	}
	s.Mutex.Unlock()
}

// WriteBytesAll sends a BinaryMessage to all connections
func (s *Server) WriteBytesAll(message []byte) {
	s.Mutex.Lock()
	for _, conn := range s.Connections {
		cloned := conn
		if cloned != nil {
			go cloned.WriteBytes(message)
		}
	}
	s.Mutex.Unlock()
}

// GetConnection can return nil
func (s *Server) GetConnection(connId ConnectionId) *Connection {
	s.Mutex.Lock()
	conn := s.Connections[connId]
	s.Mutex.Unlock()
	return conn
}
