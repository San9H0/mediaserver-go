package servers

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils/log"

	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/whep"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/dto"
)

type WebRTCServer struct {
	se  pion.SettingEngine
	me  *pion.MediaEngine
	hub *hubs.Hub
}

func NewWHEP(hub *hubs.Hub, se pion.SettingEngine) (WebRTCServer, error) {
	me := &pion.MediaEngine{}
	for mediaType, capabilities := range engines.GetWHEPRTPHeaderExtensionCapabilities() {
		for _, capa := range capabilities {
			if err := me.RegisterHeaderExtension(capa, mediaType); err != nil {
				log.Logger.Error("failed to register header extension", zap.Error(err))
			}
		}
	}

	for mediaType, capabilities := range engines.GetWebRTCCapabilities(true) {
		for _, capa := range capabilities {
			if err := me.RegisterCodec(capa, mediaType); err != nil {
				log.Logger.Error("failed to register codec", zap.Error(err))
			}
		}
	}

	return WebRTCServer{
		hub: hub,
		se:  se,
		me:  me,
	}, nil
}

func (f *WebRTCServer) StartSession(streamID string, req dto.WHEPRequest) (dto.WHEPResponse, error) {
	stream, ok := f.hub.GetStream(streamID)
	if !ok {
		return dto.WHEPResponse{}, errors.New("stream not found")
	}

	handler, ctx := whep.NewHandler(f.se, f.me)
	if err := handler.Init(ctx, stream, req.Offer); err != nil {
		return dto.WHEPResponse{}, err
	}

	sess := sessions.NewSession2[*whep.TrackContext](handler, stream)
	go func() {
		if err := sess.Run(ctx); err != nil {
			log.Logger.Error("session error", zap.Error(err))
		}
	}()

	return dto.WHEPResponse{
		Answer: handler.Answer(),
	}, nil
}
