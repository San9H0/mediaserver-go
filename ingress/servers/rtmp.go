package servers

import (
	"fmt"
	"github.com/yutopp/go-rtmp"
	"io"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
	"net"
)

type RTMPServer struct {
	hub *hubs.Hub
}

func NewRTMPServer(hub *hubs.Hub) (RTMPServer, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		return RTMPServer{}, fmt.Errorf("failed to resolve tcp address: %w", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return RTMPServer{}, fmt.Errorf("failed to listen tcp: %w", err)
	}

	rtmpServer := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			log.Logger.Info("new rtmp server onConnect")
			return conn, &rtmp.ConnConfig{
				Handler: sessions.NewRTMPSession(hub),
			}
		},
	})
	if err := rtmpServer.Serve(listener); err != nil {
		return RTMPServer{}, fmt.Errorf("failed to serve rtmp: %w", err)
	}
	return RTMPServer{
		hub: hub,
	}, nil
}

func (f *RTMPServer) StartSession(streamID string, req dto.RTMPRequest) (dto.RTMPResponse, error) {
	return dto.RTMPResponse{}, nil
}
