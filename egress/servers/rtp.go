package servers

import (
	"context"
	"errors"
	"mediaserver-go/dto"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/types"
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

	var videotrack *hubs.Track
	for _, track := range stream.Tracks() {
		if track.MediaType() == types.MediaTypeVideo {
			videotrack = track
			break
		}
	}

	fileSession, err := sessions.NewRTPSession(req.Addr, req.Port, videotrack)
	if err != nil {
		return dto.EgressRTPResponse{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	go fileSession.Run(ctx)
	return dto.EgressRTPResponse{
		PayloadType: fileSession.PayloadType(),
		SDP:         fileSession.SDP(),
	}, nil
}
