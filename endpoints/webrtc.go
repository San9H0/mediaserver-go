package endpoints

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	dto2 "mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
	"net/http"
)

type WhipHandler struct {
	server           WHIPServer
	egressFileServer EgressFileServer
}

func NewWhipHandler(server WHIPServer, egressFileServer EgressFileServer) WhipHandler {
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

	log.Logger.Debug("whip body",
		zap.String("messageType", "request"),
		zap.String("body", string(b)),
	)
	streamID := token
	resp, err := w.server.StartSession(streamID, dto2.WHIPRequest{
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
	log.Logger.Debug("whip body",
		zap.String("messageType", "response"),
		zap.String("body", string(resp.Answer)),
	)
	return nil
}

type WHEPHandler struct {
	whepServer WHEPServer
}

func NewWHEPHandler(whepServer WHEPServer) WHEPHandler {
	return WHEPHandler{
		whepServer: whepServer,
	}
}

func (w *WHEPHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}
	if c.Request().Header.Get(echo.HeaderContentType) != "application/sdp" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type")
	}
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	resp, err := w.whepServer.StartSession(token, dto2.WHEPRequest{
		Token: token,
		Offer: string(b),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Set("Content-Type", "application/sdp")
	c.Response().Header().Set("Location", "http://127.0.0.1/v1/whip/candidates")
	c.Response().WriteHeader(http.StatusCreated)
	if _, err = c.Response().Write([]byte(resp.Answer)); err != nil {
		return err
	}
	return nil
}
