package main

//import (
//	"github.com/daominah/goxiangqi/pkg/core"
//	"github.com/daominah/goxiangqi/pkg/driver/mysql"
//	"github.com/daominah/goxiangqi/pkg/driver/websocket"
//	"github.com/daominah/goxiangqi/pkg/zaplog"
//	"net/http"
//	_ "net/http/pprof"
//	"os"
//)
//
//var log = zaplog.NewLogger()
//
//func main() {
//	go func() {
//		pprofPort := os.Getenv("PPROF_PORT")
//		if pprofPort == "" {
//			log.Fatal("need to set env PPROF_PORT")
//		}
//		log.Fatal(http.ListenAndServe(pprofPort, nil))
//	}()
//
//	//readHandler := &adapter.DemoHandler{}
//	readHandler := &core.XqHandler{}
//
//	readHandler.Instrument = core.NewXqInstrument()
//
//	wsServer := websocket.NewWebsocketServer(readHandler)
//	readHandler.Writer, readHandler.Binder = wsServer, wsServer
//
//	mysqlConf, err := mysql.LoadConfig()
//	if err != nil {
//		log.Fatal("cannot mysql.LoadConfig: ", err)
//	}
//	db, err := mysql.OpenDbPool(mysqlConf)
//	if err != nil {
//		log.Fatalf("cannot open mysql db: %v", err)
//	}
//	playerRepo := mysql.MysqlPlayerRepo{DB: db}
//	readHandler.PlayerService = core.NewPlayerServiceImpl(playerRepo)
//	historyService := core.HistoryServiceImpl{
//		RatingRepo: mysql.MysqlRatingRepo{DB: db},
//		MatchRepo:  mysql.MysqlMatchRepo{DB: db},
//	}
//	readHandler.HistoryService = historyService
//
//	roomService := core.NewRoomServiceImpl()
//	roomService.Writer, roomService.Binder = wsServer, wsServer
//	roomService.HistoryService = historyService
//	go roomService.Serve()
//	readHandler.RoomService = roomService
//
//	err = wsServer.ListenAndServe()
//	if err != nil {
//		log.Fatal("cannot wsServer.ListenAndServe: ", err)
//	}
//}
