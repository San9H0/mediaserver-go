package endpoints

import (
	"github.com/labstack/echo/v4"
	"mediaserver-go/utils/dto"
	"net/http"
)

type IngressFileHandler struct {
	ingressServer IngressFileServer
}

func NewIngressFileHandler(ingressServer IngressFileServer) IngressFileHandler {
	return IngressFileHandler{
		ingressServer: ingressServer,
	}
}

func (w *IngressFileHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	var req dto.IngressFileRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	streamID := token
	resp, err := w.ingressServer.StartSession(streamID, req)
	if err != nil {
		return err
	}
	_ = resp

	c.Response().WriteHeader(http.StatusOK)
	return nil
}

type EgressFileHandler struct {
	egressServer EgressFileServer
}

func NewEgressFileHandler(egressServer EgressFileServer) EgressFileHandler {
	return EgressFileHandler{
		egressServer: egressServer,
	}
}

func (w *EgressFileHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	var req dto.EgressFileRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	streamID := token
	resp, err := w.egressServer.StartSession(streamID, req)
	if err != nil {
		return err
	}
	_ = resp

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
