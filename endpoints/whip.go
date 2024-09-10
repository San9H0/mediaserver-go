package endpoints

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"mediaserver-go/dto"
	"net/http"
	"time"
)

type WhipHandler struct {
	server           WebRTCServer
	egressFileServer EgressFileServer
}

func NewWhipHandler(server WebRTCServer, egressFileServer EgressFileServer) WhipHandler {
	return WhipHandler{
		server:           server,
		egressFileServer: egressFileServer,
	}
}

func (w *WhipHandler) Handle(c echo.Context) error {
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	token, err := getToken(c)
	if err != nil {
		return err
	}

	resp, err := w.server.StartSession(dto.WebRTCRequest{
		Token: token,
		Offer: string(b),
	})
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(time.Second)
		fmt.Println("[TESTDEBUG] egressFileServer")
		if _, err := w.egressFileServer.StartSession("WebRTCServer", dto.EgressFileRequest{}); err != nil {
			fmt.Println("[TESTDEBUG] egressFileServer error:", err)
		}
	}()

	c.Response().Header().Set("Content-Type", "application/sdp")
	c.Response().Header().Set("Location", "http://127.0.0.1/v1/whip/candidates")
	c.Response().WriteHeader(http.StatusCreated)
	if _, err = c.Response().Write([]byte(resp.Answer)); err != nil {
		return err
	}
	return nil
}
