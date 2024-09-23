package servers

import (
	"context"
	"errors"
	pion "github.com/pion/webrtc/v3"

	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/whep"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/dto"
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

	handler := whep.NewHandler(f.se)
	if err := handler.Init(stream.Sources(), req.Offer); err != nil {
		return dto.WHEPResponse{}, err
	}

	sess := sessions.NewSession[*whep.TrackContext](handler)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sess.Run(ctx)
	}()

	return dto.WHEPResponse{
		Answer: handler.Answer(),
	}, nil
}
