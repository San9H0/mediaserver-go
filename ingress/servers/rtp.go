package servers

import (
	"context"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/dto"
)

type RTPServer struct {
	hub *hubs.Hub
}

func NewRTPServer(hub *hubs.Hub) (RTPServer, error) {
	return RTPServer{
		hub: hub,
	}, nil
}

func (f *RTPServer) StartSession(streamID string, req dto.IngressRTPRequest) (dto.IngressRTPResponse, error) {
	stream := hubs.NewStream()
	f.hub.AddStream(streamID, stream)

	fileSession, err := sessions.NewRTPSession(req.Addr, req.Port, req.PayloadType, req.CodecType, stream)
	if err != nil {
		return dto.IngressRTPResponse{}, err
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fileSession.Run(ctx)
	}()

	return dto.IngressRTPResponse{}, nil
}
