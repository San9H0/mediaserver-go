package servers

import (
	"context"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/dto"
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
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	go session.Run(ctx)
	return dto.WHIPResponse{
		Answer: session.Answer(),
	}, nil
}
