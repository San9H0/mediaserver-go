package endpoints

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"mediaserver-go/dto"
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

	resp, err := w.server.StartSession(dto.WHIPRequest{
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

	resp, err := w.whepServer.StartSession(token, dto.WHEPRequest{
		Token: token,
		Offer: string(b),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	fmt.Println("[TESTDEBUG] token:", token)
	fmt.Println("[TESTDEBUG] b:", string(b))
	c.Response().Header().Set("Content-Type", "application/sdp")
	c.Response().Header().Set("Location", "http://127.0.0.1/v1/whip/candidates")
	c.Response().WriteHeader(http.StatusCreated)
	if _, err = c.Response().Write([]byte(resp.Answer)); err != nil {
		return err
	}
	return nil
}
