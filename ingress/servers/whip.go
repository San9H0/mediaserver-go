package servers

import (
	"context"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/dto"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
)

type WHIPServer struct {
	api *pion.API

	hub *hubs.Hub
}

func NewWHIP(hub *hubs.Hub, se pion.SettingEngine) (WHIPServer, error) {
	me := &pion.MediaEngine{}
	if err := me.RegisterCodec(pion.RTPCodecParameters{
		RTPCodecCapability: pion.RTPCodecCapability{
			MimeType:     pion.MimeTypeH264,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			RTCPFeedback: nil,
		},
		PayloadType: 127,
	}, pion.RTPCodecTypeVideo); err != nil {
		return WHIPServer{}, err
	}
	if err := me.RegisterCodec(pion.RTPCodecParameters{
		RTPCodecCapability: pion.RTPCodecCapability{
			MimeType:     pion.MimeTypeOpus,
			ClockRate:    48000,
			Channels:     2,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 96,
	}, pion.RTPCodecTypeAudio); err != nil {
		return WHIPServer{}, err
	}
	api := pion.NewAPI(pion.WithSettingEngine(se), pion.WithMediaEngine(me))
	return WHIPServer{
		api: api,
		hub: hub,
	}, nil
}

func (w *WHIPServer) StartSession(req dto.WHIPRequest) (dto.WHIPResponse, error) {
	streamID := req.Token
	stream := hubs.NewStream()
	w.hub.AddStream(streamID, stream)

	session, err := sessions.NewWHIPSession(req.Offer, req.Token, w.api, stream)
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
