package servers

import (
	"context"
	"errors"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/rtp"
	"mediaserver-go/hubs"
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

func (f *RTPServer) StartSession(streamID string, req dto.EgressRTPRequest) (dto.EgressRTPResponse, error) {
	stream, ok := f.hub.GetStream(streamID)
	if !ok {
		return dto.EgressRTPResponse{}, errors.New("stream not found")
	}

	filteredSourceTracks, err := filterMediaTypesInStream(stream, req.MediaTypes)
	if err != nil {
		return dto.EgressRTPResponse{}, err
	}

	handler := rtp.NewHandler(req.Addr, req.Port)
	if err := handler.Init(context.Background(), filteredSourceTracks); err != nil {
		return dto.EgressRTPResponse{}, err
	}

	sess := sessions.NewSession[*rtp.TrackContext](handler)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sess.Run(ctx)
	}()

	return dto.EgressRTPResponse{
		SDP: handler.SDP(),
	}, nil
}
