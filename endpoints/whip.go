package endpoints

import (
	"github.com/labstack/echo/v4"
	"io"
	"mediaserver-go/dto"
	"net/http"
)

type WhipHandler struct {
	server WebRTCServer
}

func NewWhipHandler(server WebRTCServer) WhipHandler {
	return WhipHandler{
		server: server,
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

	c.Response().Header().Set("Content-Type", "application/sdp")
	c.Response().Header().Set("Location", "http://127.0.0.1/v1/whip/candidates")
	c.Response().WriteHeader(http.StatusCreated)
	if _, err = c.Response().Write([]byte(resp.Answer)); err != nil {
		return err
	}
	return nil
}
