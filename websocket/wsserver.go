package websocket

import (
	"net/http"
	"sync"

	"github.com/daominah/gomicrokit/httpsvr"
	"github.com/daominah/gomicrokit/log"
	goraws "github.com/gorilla/websocket"
)

type Server struct {
	listeningPort    string
	wsPath           string
	HttpRouter       *httpsvr.Server
	wsOnReadHandler  OnReadHandler
	wsOnCloseHandler OnCloseHandler
	Connections      map[ConnectionId]*Connection
	Mutex            sync.Mutex
}

func NewServer(listeningPort string, wsPath string,
	wsOnReadHandler OnReadHandler, wsOnCloseHandler OnCloseHandler) *Server {
	if wsOnReadHandler == nil {
		wsOnReadHandler = &emptyHandler{}
	}
	if wsOnCloseHandler == nil {
		wsOnCloseHandler = &emptyHandler{}
	}
	s := &Server{
		listeningPort:    listeningPort,
		wsPath:           wsPath,
		Connections:      make(map[ConnectionId]*Connection),
		HttpRouter:       httpsvr.NewServer(),
		wsOnReadHandler:  wsOnReadHandler,
		wsOnCloseHandler: wsOnCloseHandler,
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
		conn := NewConnection(goraConn, s.wsOnReadHandler)
		s.Connections[conn.id] = conn
		s.Mutex.Unlock()
		log.Condf(LOG, "%v connected", conn.id)
		go func() {
			<-conn.ClosedChan
			s.Mutex.Lock()
			delete(s.Connections, conn.id)
			s.Mutex.Unlock()
			log.Condf(LOG, "%v disconnected", conn.id)
			s.wsOnCloseHandler.OnClose(conn.id)
		}()
	}
}

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

func (s *Server) WriteBytes(connId ConnectionId, message []byte) {
	s.Mutex.Lock()
	conn := s.Connections[connId]
	s.Mutex.Unlock()
	if conn != nil {
		conn.WriteBytes(message)
	}
}

func (s *Server) WriteAll(message string) {
	s.Mutex.Lock()
	for _, conn := range s.Connections {
		if conn != nil {
			go conn.Write(message)
		}
	}
	s.Mutex.Unlock()
}

func (s *Server) WriteBytesAll(message []byte) {
	s.Mutex.Lock()
	for _, conn := range s.Connections {
		if conn != nil {
			go conn.WriteBytes(message)
		}
	}
	s.Mutex.Unlock()
}
