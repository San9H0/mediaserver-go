package endpoints

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/pion/webrtc/v3"
	"mediaserver-go/dto"
	"time"
)

type WebRTCServer interface {
	StartSession(request dto.WebRTCRequest) (dto.WebRTCResponse, error)
}

type Request struct {
	Token string
	Offer string
}

func Initialize(s WebRTCServer) *echo.Echo {
	// Create a new Echo instance
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Println("err:", err)
	}

	// Apply the request logger middleware
	e.Use(RequestLogger)
	e.GET("/v1/whip/candidates", func(c echo.Context) error {
		fmt.Println("candidates")
		return nil
	})

	whipHandler := NewWhipHandler(s)
	e.POST("/v1/whip", whipHandler.Handle)

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
