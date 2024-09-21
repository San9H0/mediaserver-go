package endpoints

import (
	"github.com/labstack/echo/v4"
	"mediaserver-go/utils/dto"
	"net/http"
)

type ImageServer interface {
	StartSession(streamID string, request dto.ImagesRequest) (dto.ImagesResponse, error)
}

type ImagesHandler struct {
	imageServer ImageServer
}

func NewImagesHandler(imageServer ImageServer) ImagesHandler {
	return ImagesHandler{
		imageServer: imageServer,
	}
}

func (i *ImagesHandler) Register(e *echo.Echo) {
	//e.POST("/v1/hls", i.Handle)
}

func (i *ImagesHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	var req dto.ImagesRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	streamID := token
	resp, err := i.imageServer.StartSession(streamID, req)
	if err != nil {
		return err
	}
	_ = resp

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
