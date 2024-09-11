package servers

import (
	"context"
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/dto"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/hubs"
)

type WebRTCServer struct {
	se  pion.SettingEngine
	hub *hubs.Hub
}

func NewWHEP(hub *hubs.Hub, se pion.SettingEngine) (WebRTCServer, error) {
	return WebRTCServer{
		hub: hub,
		se:  se,
	}, nil
}

func (f *WebRTCServer) StartSession(streamID string, req dto.WHEPRequest) (dto.WHEPResponse, error) {
	stream, ok := f.hub.GetStream(streamID)
	if !ok {
		return dto.WHEPResponse{}, errors.New("stream not found")
	}

	whepSession, err := sessions.NewWHEPSession(req.Offer, req.Token, f.se, stream.Tracks())
	if err != nil {
		return dto.WHEPResponse{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	go whepSession.Run(ctx)

	return dto.WHEPResponse{
		Answer: whepSession.Answer(),
	}, nil
}
