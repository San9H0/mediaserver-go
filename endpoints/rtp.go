package endpoints

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"mediaserver-go/dto"
	"net/http"
)

type RTPHandler struct {
	ingressRTPServer IngressRTPServer
	egressRTPServer  EgressRTPServer
}

func NewRTPHandler(ingressRTPServer IngressRTPServer, egressRTPServer EgressRTPServer) RTPHandler {
	return RTPHandler{
		ingressRTPServer: ingressRTPServer,
		egressRTPServer:  egressRTPServer,
	}
}

func (w *RTPHandler) HandleEgress(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	var req dto.EgressRTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	req.Token = token
	resp, err := w.egressRTPServer.StartSession(token, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	c.Response().WriteHeader(http.StatusOK)
	if _, err = c.Response().Write(b); err != nil {
		return err
	}
	return nil
}

func (w *RTPHandler) HandleIngress(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	var req dto.IngressRTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	req.Token = token
	resp, err := w.ingressRTPServer.StartSession(token, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	_ = resp

	c.Response().WriteHeader(http.StatusOK)
	if _, err = c.Response().Write(nil); err != nil {
		return err
	}
	return nil
}
