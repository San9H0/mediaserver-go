package servers

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
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

	fileSession, err := sessions.NewRTPSession(req.Addr, req.Port, filteredSourceTracks)
	if err != nil {
		return dto.EgressRTPResponse{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	go func() {
		if err := fileSession.Run(ctx); err != nil {
			log.Logger.Warn("file session error", zap.Error(err))
		}
	}()
	return dto.EgressRTPResponse{
		PayloadType: fileSession.PayloadType(),
		SDP:         fileSession.SDP(),
	}, nil
}
