package whep

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/factory"
	"mediaserver-go/codecs/opus"
	"mediaserver-go/egress/sessions/whep/playoutdelay"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"strings"

	"mediaserver-go/hubs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type TrackContext struct {
	track hubs.Track
}

type Handler struct {
	cancel             context.CancelFunc
	se                 pion.SettingEngine
	me                 *pion.MediaEngine
	remoteTrackHandler map[types.MediaType]*RemoteTrackHandler
	rtcpCh             chan rtcp.Packet

	api               *pion.API
	pc                *pion.PeerConnection
	onConnectionState chan pion.PeerConnectionState
}

func NewHandler(se pion.SettingEngine, me *pion.MediaEngine) (*Handler, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Handler{
		cancel:             cancel,
		se:                 se,
		me:                 me,
		rtcpCh:             make(chan rtcp.Packet, 10),
		remoteTrackHandler: make(map[types.MediaType]*RemoteTrackHandler),
	}, ctx
}

func (h *Handler) PreferredCodec(originalCodec codecs.Codec) codecs.Codec {
	if originalCodec.MediaType() == types.MediaTypeVideo {
		return originalCodec
	}
	if _, err := originalCodec.WebRTCCodecCapability(); err == nil {
		return originalCodec
	}
	return opus.NewOpus(opus.NewConfig(opus.Parameters{
		Channels:     2,
		SampleRate:   48000,
		SampleFormat: int(avutil.AV_SAMPLE_FMT_FLT),
	}))
}

func (h *Handler) Answer() string {
	return h.pc.LocalDescription().SDP
}

func (h *Handler) Init(ctx context.Context, stream *hubs.Stream, offer string) error {
	api := pion.NewAPI(pion.WithMediaEngine(h.me), pion.WithSettingEngine(h.se))

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return err
	}

	streamID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	localTrackMap := make(map[types.MediaType]*pion.TrackLocalStaticRTP)
	for mediaType, codec := range stream.GetCodecs() {
		trackID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		availableCodec := h.PreferredCodec(codec)
		webrtcCodecCapability, err := availableCodec.WebRTCCodecCapability()
		if err != nil {
			return err
		}
		localTrack, err := pion.NewTrackLocalStaticRTP(webrtcCodecCapability, trackID.String(), streamID.String())
		if err != nil {
			return err
		}
		localTrackMap[mediaType] = localTrack
		transceiver, err := pc.AddTransceiverFromTrack(localTrack)
		if err != nil {
			return err
		}
		if transceiver.Sender() == nil {
			return fmt.Errorf("transceiver sender is nil")
		}
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		fmt.Println("[TESTDEBUG] SetRemoteDescription err:", err)
		return err
	}

	candidateCh := make(chan *pion.ICECandidate, 10)
	pc.OnICECandidate(func(candidate *pion.ICECandidate) {
		if candidate == nil {
			close(candidateCh)
			return
		}
		candidateCh <- candidate
	})
	pc.OnConnectionStateChange(func(connectionState pion.PeerConnectionState) {
		log.Logger.Info("connection state changed", zap.String("state", connectionState.String()))
		switch connectionState {
		case pion.PeerConnectionStateClosed, pion.PeerConnectionStateFailed, pion.PeerConnectionStateDisconnected:
			h.cancel()
		}
	})
	pc.OnTrack(func(remote *pion.TrackRemote, receiver *pion.RTPReceiver) {
	})

	sd, err := pc.CreateAnswer(&pion.AnswerOptions{})
	if err != nil {
		return err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return err
	}

	for range candidateCh {
	}

	for _, transceiver := range pc.GetTransceivers() {
		mediaType := types.NewMediaType(transceiver.Kind().String())
		codec := stream.GetCodecs()[mediaType]
		availableCodec := h.PreferredCodec(codec)
		packetizer, remoteRTXHandler, err := getPacketizerAndRemoteRTX(transceiver.Sender(), availableCodec.MimeType())
		if err != nil {
			return err
		}

		var playoutDelayHandler *playoutdelay.Handler
		var getExtensions []func() (int, []byte, bool)
		for _, ext := range transceiver.Sender().GetParameters().HeaderExtensions {
			switch ext.URI {
			case "http://www.webrtc.org/experiments/rtp-hdrext/playout-delay":
				playoutDelayHandler = playoutdelay.NewHandler(ext.ID, true)
				getExtensions = append(getExtensions, playoutDelayHandler.GetPayload)
			}
		}

		stats := NewStats()

		remoteTrackHandler := NewRemoteTrackHandler(Args{
			mediaType:              mediaType,
			localTrack:             localTrackMap[mediaType],
			transceiver:            transceiver,
			sender:                 transceiver.Sender(),
			stats:                  stats,
			packetizer:             packetizer,
			remoteRTXHandler:       remoteRTXHandler,
			playoutDelayHandler:    playoutDelayHandler,
			getExtensions:          getExtensions,
			adaptiveBitrateHandler: NewABSHandler(stats),
			pc:                     pc,
		})
		go remoteTrackHandler.Run(ctx)

		h.remoteTrackHandler[mediaType] = remoteTrackHandler
	}

	h.api = api
	h.pc = pc
	//log.Logger.Info("whep negotiated end", zap.Int("negotiated", len(negotidated)))
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	h.cancel()
	h.pc.Close()

	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track hubs.Track) (*TrackContext, error) {
	codec := track.GetCodec()
	mediaType := codec.MediaType()
	h.remoteTrackHandler[mediaType].adaptiveBitrateHandler.SetMaxSpatialLayer(track.RID())
	return &TrackContext{
		track: track,
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, unit units.Unit, rid string) error {
	remoteHandler := h.remoteTrackHandler[types.MediaTypeVideo]
	if rid == "" {
		return remoteHandler.onVideo(ctx, trackCtx.track, unit, rid)
	}

	if !remoteHandler.adaptiveBitrateHandler.CanSendSpatialLayer(rid, unit) {
		return nil
	}

	remoteHandler.packetCh <- Packet{
		track: trackCtx.track,
		unit:  unit,
		rid:   rid,
	}
	return nil
}

func (h *Handler) OnAudio(ctx context.Context, _ *TrackContext, unit units.Unit, rid string) error {
	remoteHandler := h.remoteTrackHandler[types.MediaTypeAudio]

	return remoteHandler.onAudio(ctx, nil, unit)
}

func getPacketizerAndRemoteRTX(sender *pion.RTPSender, mimeType string) (rtp.Packetizer, *remoteRTX, error) {
	oritginalPayloadType := uint8(0)
	clockRate := uint32(0)
	ssrc := uint32(sender.GetParameters().Encodings[0].SSRC)
	for _, codec := range sender.GetParameters().Codecs {
		if codec.MimeType != mimeType {
			continue
		}
		oritginalPayloadType = uint8(codec.PayloadType)
		clockRate = codec.ClockRate
	}

	base, err := factory.NewBase(mimeType)
	if err != nil {
		return nil, nil, err
	}

	packetizer, err := base.RTPPacketizer(oritginalPayloadType, ssrc, clockRate)
	if err != nil {
		return nil, nil, err
	}

	return packetizer, getRemoteRTX(sender, oritginalPayloadType), nil
}

func getRemoteRTX(sender *pion.RTPSender, originalPayloadType uint8) *remoteRTX {
	pt := uint8(0)
	for _, codec := range sender.GetParameters().Codecs {
		if codec.MimeType != "video/rtx" {
			continue
		}
		if !strings.Contains(codec.SDPFmtpLine, fmt.Sprintf("apt=%d", originalPayloadType)) {
			continue
		}
		pt = uint8(codec.PayloadType)
	}
	if pt == 0 {
		return nil
	}
	return &remoteRTX{
		sequenceNumber: 0,
		payloadType:    pt,
	}
}
