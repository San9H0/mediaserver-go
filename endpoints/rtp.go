package endpoints

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
	"net/http"
)

type IngressRTPHandler struct {
	ingressRTPServer IngressRTPServer
}

func NewIngressRTPHandler(ingressRTPServer IngressRTPServer) IngressRTPHandler {
	return IngressRTPHandler{
		ingressRTPServer: ingressRTPServer,
	}
}

func (i *IngressRTPHandler) HandleIngress(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	var req dto.IngressRTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	streamID := token
	resp, err := i.ingressRTPServer.StartSession(streamID, req)
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

type EgressRTPPHandler struct {
	egressRTPServer EgressRTPServer
}

func NewEgressRTPHandler(egressRTPServer EgressRTPServer) EgressRTPPHandler {
	return EgressRTPPHandler{
		egressRTPServer: egressRTPServer,
	}
}

func (w *EgressRTPPHandler) HandleEgress(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	var req dto.EgressRTPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	log.Logger.Debug("egress rtp body",
		zap.String("messageType", "request"),
		zap.Any("body", req),
	)

	streamID := token
	resp, err := w.egressRTPServer.StartSession(streamID, req)
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

	log.Logger.Debug("egress rtp body",
		zap.String("messageType", "response"),
		zap.Any("body", string(b)),
	)
	return nil
}
