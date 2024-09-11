package endpoints

import (
	"github.com/labstack/echo/v4"
	"mediaserver-go/dto"
	"net/http"
)

type FileHandler struct {
	ingressServer IngressFileServer
	egressServer  EgressFileServer
}

func NewFileHandler(ingressServer IngressFileServer, egressServer EgressFileServer) FileHandler {
	return FileHandler{
		ingressServer: ingressServer,
		egressServer:  egressServer,
	}
}

func (w *FileHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	var req dto.IngressFileRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.Token = token
	resp, err := w.ingressServer.StartSession(req)
	if err != nil {
		return err
	}
	_ = resp
	//r, err := w.egressServer.StartSession("FileServerID", dto.EgressFileRequest{
	//	Token: token,
	//})
	//_ = r
	//if err != nil {
	//	return err
	//}

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
