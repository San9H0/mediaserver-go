package endpoints

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/pion/webrtc/v3"
	"mediaserver-go/dto"
	"time"
)

type WHIPServer interface {
	StartSession(request dto.WHIPRequest) (dto.WHIPResponse, error)
}

type IngressFileServer interface {
	StartSession(request dto.IngressFileRequest) (dto.IngressFileResponse, error)
}

type WHEPServer interface {
	StartSession(streamID string, request dto.WHEPRequest) (dto.WHEPResponse, error)
}

type EgressFileServer interface {
	StartSession(streamID string, request dto.EgressFileRequest) (dto.EgressFileResponse, error)
}

type IngressRTPServer interface {
	StartSession(streamID string, request dto.IngressRTPRequest) (dto.IngressRTPResponse, error)
}

type EgressRTPServer interface {
	StartSession(streamID string, request dto.EgressRTPRequest) (dto.EgressRTPResponse, error)
}

type Request struct {
	Token string
	Offer string
}

func Initialize(
	whipServer WHIPServer,
	ingressFileServer IngressFileServer,
	whepServer WHEPServer,
	egressFileServer EgressFileServer,
	ingressRTPServer IngressRTPServer,
	egressRTPServer EgressRTPServer) *echo.Echo {
	// Create a new Echo instance
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Println("err:", err)
	}

	e.Use(RequestLogger)

	e.GET("/v1/whip/candidates", func(c echo.Context) error {
		fmt.Println("candidates")
		return nil
	})

	whipHandler := NewWhipHandler(whipServer, egressFileServer)
	e.POST("/v1/whip", whipHandler.Handle)

	whepHandler := NewWHEPHandler(whepServer)
	e.POST("/v1/whep", whepHandler.Handle)

	fileHandler := NewFileHandler(ingressFileServer, egressFileServer)
	e.POST("/v1/files", fileHandler.Handle)

	egressRTPHandler := NewRTPHandler(ingressRTPServer, egressRTPServer)
	e.POST("/v1/rtp/out", egressRTPHandler.HandleEgress)
	e.POST("/v1/rtp/in", egressRTPHandler.HandleIngress)

	return e
}

func getToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}
	return "", errors.New("no token")
}

// RequestLogger logs all requests including 404s
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Log request details
		start := time.Now()
		err := next(c)
		stop := time.Now()

		method := c.Request().Method
		uri := c.Request().RequestURI
		status := c.Response().Status

		fmt.Printf("[%s] %s %s %d %s\n", start.Format(time.RFC3339), method, uri, status, stop.Sub(start))

		return err
	}
}
