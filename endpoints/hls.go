package endpoints

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"mediaserver-go/utils/dto"
)

type HLSHandler struct {
	hlsServer HLSServer
}

func NewHLSHandler(hlsServer HLSServer) HLSHandler {
	return HLSHandler{
		hlsServer: hlsServer,
	}
}

func (h *HLSHandler) Register(e *echo.Echo) {
	e.POST("/v1/hls", h.Handle)
	e.GET("/v1/llhls/:streamID/:target", func(c echo.Context) error {
		streamID, target := c.Param("streamID"), c.Param("target")
		if streamID == "" || target == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid parameters")
		}
		handle, err := h.hlsServer.GetHLSStream(streamID)
		fmt.Println("err:", err)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		switch target {
		case "index.m3u8":
			b, err := handle.GetMasterM3U8()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "application/vnd.apple.mpegurl", b)
		case "video.m3u8":
			medisSN, part := c.QueryParam("_HLS_msn"), c.QueryParam("_HLS_part")
			b, err := handle.GetMediaM3U8LLHLS(medisSN, part)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "application/vnd.apple.mpegurl", b)
		case "init.mp4":
			b, err := handle.GetPayload(target)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "video/mp4", b)
		default:
			b, err := handle.GetPayload(target)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "video/mp4", b)
		}
	})
	e.GET("/v1/hls/:streamID/:target", func(c echo.Context) error {
		streamID, target := c.Param("streamID"), c.Param("target")
		if streamID == "" || target == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid parameters")
		}
		handle, err := h.hlsServer.GetHLSStream(streamID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		switch target {
		case "index.m3u8":
			b, err := handle.GetMasterM3U8()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "application/vnd.apple.mpegurl", b)
		case "video.m3u8":
			b, err := handle.GetMediaM3U8HLS()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "application/vnd.apple.mpegurl", b)
		case "init.mp4":
			b, err := handle.GetPayload(target)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "video/mp4", b)
		default:
			b, err := handle.GetPayload(target)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			return c.Blob(http.StatusOK, "video/mp4", b)
		}
	})
}

func (h *HLSHandler) Handle(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	var req dto.HLSRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	streamID := token
	resp, err := h.hlsServer.StartSession(streamID, req)
	if err != nil {
		return err
	}
	_ = resp

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
