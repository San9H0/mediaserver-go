package servers

import (
	"context"
	"github.com/pion/ice/v2"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/dto"
	"mediaserver-go/ingress/sessions"
)

type WebRTCServer struct {
	api *pion.API
}

func NewWebRTC() (WebRTCServer, error) {
	se := pion.SettingEngine{}
	if err := se.SetEphemeralUDPPortRange(10000, 20000); err != nil {
		return WebRTCServer{}, err
	}
	se.SetIncludeLoopbackCandidate(true)
	se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	//se.SetNAT1To1IPs([]string{"127.0.0.1"}, webrtc.ICECandidateTypeHost)
	se.SetLite(true)

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
		return WebRTCServer{}, err
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
		return WebRTCServer{}, err
	}
	api := pion.NewAPI(pion.WithSettingEngine(se), pion.WithMediaEngine(me))
	return WebRTCServer{
		api: api,
	}, nil
}

func (w *WebRTCServer) StartSession(req dto.WebRTCRequest) (dto.WebRTCResponse, error) {
	session, err := sessions.NewSession(req.Offer, req.Token, w.api)
	if err != nil {
		return dto.WebRTCResponse{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	go session.Run(ctx)
	return dto.WebRTCResponse{
		Answer: session.Answer(),
	}, nil
}
