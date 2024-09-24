package servers

import (
	"context"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
)

type WHIPServer struct {
	api *pion.API

	hub *hubs.Hub
}

func NewWHIP(hub *hubs.Hub, se pion.SettingEngine) (WHIPServer, error) {
	me := &pion.MediaEngine{}
	for kind, capabilities := range engines.GetWebRTCCapabilities() {
		for _, capability := range capabilities {
			if err := me.RegisterCodec(capability, kind); err != nil {
			}
		}
	}
	api := pion.NewAPI(pion.WithSettingEngine(se), pion.WithMediaEngine(me))
	return WHIPServer{
		api: api,
		hub: hub,
	}, nil
}

func (w *WHIPServer) StartSession(streamID string, req dto.WHIPRequest) (dto.WHIPResponse, error) {
	stream := hubs.NewStream()
	w.hub.AddStream(streamID, stream)

	session, err := sessions.NewWHIPSession(req.Offer, streamID, w.api, stream)
	if err != nil {
		return dto.WHIPResponse{}, err
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := session.Run(ctx); err != nil {
			log.Logger.Error("WHIP session error", zap.Error(err))
		}
	}()

	return dto.WHIPResponse{
		Answer: session.Answer(),
	}, nil
}
