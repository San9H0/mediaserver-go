package endpoints

import (
	"errors"
	"fmt"
	"mediaserver-go/egress/servers"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/pion/webrtc/v3"
	"go.uber.org/zap"

	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
)

type WHIPServer interface {
	StartSession(streamID string, request dto.WHIPRequest) (dto.WHIPResponse, error)
}
type IngressRTPServer interface {
	StartSession(streamID string, request dto.IngressRTPRequest) (dto.IngressRTPResponse, error)
}
type IngressFileServer interface {
	StartSession(streamID string, request dto.IngressFileRequest) (dto.IngressFileResponse, error)
}

type WHEPServer interface {
	StartSession(streamID string, request dto.WHEPRequest) (dto.WHEPResponse, error)
}
type EgressFileServer interface {
	StartSession(streamID string, request dto.EgressFileRequest) (dto.EgressFileResponse, error)
}
type EgressRTPServer interface {
	StartSession(streamID string, request dto.EgressRTPRequest) (dto.EgressRTPResponse, error)
}
type HLSServer interface {
	StartSession(streamID string, request dto.HLSRequest) (dto.HLSResponse, error)
	GetHLSStream(streamID string) (*servers.HLSHandler, error)
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
	egressRTPServer EgressRTPServer,
	hlsServer HLSServer,
	imageServer ImageServer,
) *echo.Echo {
	// Create a new Echo instance
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {

		log.Logger.Warn("http error",
			zap.String("request_url", c.Request().URL.String()),
			zap.Error(err),
		)
	}

	e.Use(RequestLogger)

	e.GET("/v1/whip/candidates", func(c echo.Context) error {
		fmt.Println("candidates")
		return nil
	})

	whipHandler := NewWhipHandler(whipServer, egressFileServer)
	ingressFileHandler := NewIngressFileHandler(ingressFileServer)
	ingressRTPHandler := NewIngressRTPHandler(ingressRTPServer)
	e.POST("/v1/whip", whipHandler.Handle)
	e.POST("/v1/ingress/files", ingressFileHandler.Handle)
	e.POST("/v1/ingress/rtp", ingressRTPHandler.HandleIngress)

	whepHandler := NewWHEPHandler(whepServer)
	egressFileHandler := NewEgressFileHandler(egressFileServer)
	egressRTPHandler := NewEgressRTPHandler(egressRTPServer)
	hlsHandler := NewHLSHandler(hlsServer)

	e.POST("/v1/whep", whepHandler.Handle)
	e.POST("/v1/egress/files", egressFileHandler.Handle)
	e.POST("/v1/egress/rtp", egressRTPHandler.HandleEgress)

	hlsHandler.Register(e)

	imageHandler := NewImagesHandler(imageServer)
	imageHandler.Register(e)

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
