package servers

import (
	"fmt"
	"github.com/yutopp/go-rtmp"
	"io"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/log"
	"net"
)

type RTMPServer struct {
	rtmpServer *rtmp.Server
	hub        *hubs.Hub
}

func NewRTMPServer(hub *hubs.Hub) (RTMPServer, error) {
	rtmpServer := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			log.Logger.Info("new rtmp server onConnect")
			return conn, &rtmp.ConnConfig{
				Handler: sessions.NewRTMPSession(hub),
			}
		},
	})
	return RTMPServer{
		rtmpServer: rtmpServer,
		hub:        hub,
	}, nil
}

func (r *RTMPServer) Start(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve tcp address: %w", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %w", err)
	}
	return r.rtmpServer.Serve(listener)
}
