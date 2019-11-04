package websocket
//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/daominah/goxiangqi/pkg/core"
//	"github.com/daominah/goxiangqi/pkg/maths"
//	"github.com/gorilla/websocket"
//	"github.com/julienschmidt/httprouter"
//	"net/http"
//	"os"
//	"sync"
//	"time"
//)
//
//type Server struct {
//	connections  map[core.ConnectionId]*Connection
//	playersConns map[core.PlayerId]map[core.ConnectionId]bool
//	connsPlayer  map[core.ConnectionId]core.PlayerId
//	mutex        sync.Mutex         `json:"-"`
//	httpRouter   *httprouter.Router `json:"-"`
//	handler      core.Handler       `json:"-"`
//}
//
//func NewWebsocketServer(handler core.Handler) *Server {
//	s := &Server{handler: handler}
//	s.connections = make(map[core.ConnectionId]*Connection)
//	s.playersConns = make(map[core.PlayerId]map[core.ConnectionId]bool)
//	s.connsPlayer = make(map[core.ConnectionId]core.PlayerId)
//	s.httpRouter = httprouter.New()
//	s.httpRouter.HandlerFunc(http.MethodGet, "/", s.handleUpgradeWs())
//	s.httpRouter.HandlerFunc(http.MethodGet, "/debug/connections",
//		s.handleDebugConnections())
//	return s
//}
//
//func (s *Server) handleUpgradeWs() http.HandlerFunc {
//	upgrader := websocket.Upgrader{
//		ReadBufferSize:  8192,
//		WriteBufferSize: 8192,
//		CheckOrigin:     func(r *http.Request) bool { return true },
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		wsConn, err := upgrader.Upgrade(w, r, nil)
//		if err != nil {
//			log.Info("cannot upgrader.Upgrade: ", r.RemoteAddr, err)
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		connId := core.ConnectionId(maths.GenUUID())
//		conn := &Connection{
//			Id:             connId,
//			conn:           wsConn,
//			CreateAt:       time.Now(),
//			WriteChan:      make(chan []byte),
//			ReadHandler:    s.handler,
//			Server:         s,
//			disconnectChan: make(chan bool),
//		}
//		log.Debugf("new connection %v: %v", conn.conn.RemoteAddr(), connId)
//		s.mutex.Lock()
//		s.connections[conn.Id] = conn
//		s.mutex.Unlock()
//		go conn.ReadPump()
//		go conn.WritePump()
//		go conn.DisconnectCallback()
//	}
//}
//
//func (s *Server) ListenAndServe() error {
//	listeningPort := os.Getenv("WEBSOCKET_PORT")
//	if listeningPort == "" {
//		return errors.New("need to set WEBSOCKET_PORT env")
//	}
//	log.Info("starting websocket server on port ", listeningPort)
//	return http.ListenAndServe(listeningPort, s.httpRouter)
//}
//
//func (s *Server) handleDebugConnections() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		s.mutex.Lock()
//		nConns := len(s.connections)
//		b1, err1 := json.MarshalIndent(s.connections, "", "    ")
//		b2, err2 := json.MarshalIndent(s.playersConns, "", "    ")
//		b3, err3 := json.MarshalIndent(s.connsPlayer, "", "    ")
//		s.mutex.Unlock()
//		if (err1 != nil) || (err2 != nil) || (err3 != nil) {
//			http.Error(w, fmt.Sprintf("cannot json marshal: %v, %v, %v",
//				err1, err2, err3), http.StatusInternalServerError)
//		}
//		w.Write([]byte(fmt.Sprintf(`nConns: %v
//
//connections  map[core.ConnectionId]*Connection:
//    %v
//
//playersConns map[core.PlayerId]map[core.ConnectionId]bool:
//    %v
//
//connsPlayer  map[core.ConnectionId]core.PlayerId:
//    %v
//`, nConns, string(b1), string(b2), string(b3))))
//	}
//}
//
//func (s *Server) WriteToConnection(connId core.ConnectionId, msg string) {
//	s.mutex.Lock()
//	conn := s.connections[connId]
//	s.mutex.Unlock()
//	if conn != nil {
//		conn.Write(msg)
//	}
//}
//
//func (s *Server) WriteToPlayer(playerId core.PlayerId, msg string) {
//	conns := make([]*Connection, 0)
//	s.mutex.Lock()
//	for connId := range s.playersConns[playerId] {
//		conns = append(conns, s.connections[connId])
//	}
//	s.mutex.Unlock()
//	for _, conn := range conns {
//		if conn != nil {
//			conn.Write(msg)
//		}
//	}
//}
//
//func (s *Server) Bind(connId core.ConnectionId, playerId core.PlayerId) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	old, ok := s.connsPlayer[connId]
//	if old == playerId {
//		return
//	}
//	if ok {
//		delete(s.playersConns[old], connId)
//		if len(s.playersConns[old]) == 0 {
//			delete(s.playersConns, old)
//		}
//	}
//	//
//	s.connsPlayer[connId] = playerId
//	if _, ok := s.playersConns[playerId]; !ok {
//		s.playersConns[playerId] = make(map[core.ConnectionId]bool)
//	}
//	s.playersConns[playerId][connId] = true
//}
//
//func (s *Server) Unbind(connId core.ConnectionId) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	if old, ok := s.connsPlayer[connId]; ok {
//		delete(s.playersConns[old], connId)
//		if len(s.playersConns[old]) == 0 {
//			delete(s.playersConns, old)
//		}
//	}
//	delete(s.connsPlayer, connId)
//}
//
//// delete a disconnected connection
//func (s *Server) delete(connId core.ConnectionId) {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//	delete(s.connections, connId)
//}
//
//func (s *Server) CheckOffline(pid core.PlayerId) bool {
//	s.mutex.Lock()
//	defer s.mutex.Unlock()
//
//	var isOffline bool
//	isOffline = len(s.playersConns[pid]) == 0
//	return isOffline
//}
