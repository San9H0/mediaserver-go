package endpoints

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"mediaserver-go/peertopeer"
	"time"

	"net/http"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader

	server  *peertopeer.Server
	handler *peertopeer.ServiceHandler
}

func NewWebSocketHandler() WebSocketHandler {
	// WebSocket 업그레이더
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // CORS 설정: 모든 출처 허용
		},
	}
	return WebSocketHandler{
		upgrader: upgrader,
		server:   peertopeer.NewServer(),
		handler:  peertopeer.NewMessageHandler(),
	}
}

func (w *WebSocketHandler) Handle(c echo.Context) error {
	fmt.Println("WebSocketHandler.Handle")

	conn, err := w.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	session := peertopeer.NewSession(conn, w.handler)
	w.server.AddSession(session)

	sess := w.server.GetSessions()
	if len(sess) == 2 {
		go func() {
			time.Sleep(time.Second)
			w.handler.AddPair(sess[0], sess[1])
			sess[0].Write(context.Background(), peertopeer.Header{Type: "startOffer"})
			//sess[1].Write(context.Background(), []byte("startAnswer"))
		}()
	}

	defer w.server.RemoveSession(session)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return session.Run(ctx)
}
