package whep

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"

	"mediaserver-go/hubs"
	hubcodecs "mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/ntp"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type TrackContext struct {
	sourceTrack *hubs.Track
	localTrack  *pion.TrackLocalStaticRTP
	sender      *pion.RTPSender
	packetizer  rtp.Packetizer
	buf         []byte
	stats       Stats
}

type Handler struct {
	se          pion.SettingEngine
	localTracks map[*hubs.Track]*pion.TrackLocalStaticRTP

	api               *pion.API
	pc                *pion.PeerConnection
	onConnectionState chan pion.PeerConnectionState
	negotidated       []*hubs.Track
}

func NewHandler(se pion.SettingEngine) *Handler {
	return &Handler{
		se:          se,
		localTracks: make(map[*hubs.Track]*pion.TrackLocalStaticRTP),
	}
}

func (h *Handler) NegotiatedTracks() []*hubs.Track {
	ret := make([]*hubs.Track, 0, len(h.negotidated))
	return append(ret, h.negotidated...)
}

func (h *Handler) Answer() string {
	return h.pc.LocalDescription().SDP
}

func (h *Handler) Init(tracks []*hubs.Track, offer string) error {
	onConnectionState := make(chan pion.PeerConnectionState, 10)

	var negotidated []*hubs.Track
	me := &pion.MediaEngine{}
	for _, track := range tracks {
		switch track.MediaType() {
		case types.MediaTypeVideo:
			videoCodec, err := track.VideoCodec()
			if err != nil {
				return err
			}
			webrtcCodecCapability, err := videoCodec.WebRTCCodecCapability()
			if err != nil {
				return err
			}

			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: webrtcCodecCapability,
				PayloadType:        127,
			}, pion.RTPCodecTypeVideo); err != nil {
				return err
			}

			negotidated = append(negotidated, track)
		case types.MediaTypeAudio:
			audioCodec, err := track.AudioCodec()
			if err != nil {
				return err
			}
			webrtcCodecCapability, err := audioCodec.WebRTCCodecCapability()
			if err != nil {
				return err
			}
			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: webrtcCodecCapability,
				PayloadType:        111,
			}, pion.RTPCodecTypeAudio); err != nil {
				return err
			}

			negotidated = append(negotidated, track)
		}
	}

	fmt.Println("negotidated:", len(negotidated))
	api := pion.NewAPI(pion.WithMediaEngine(me), pion.WithSettingEngine(h.se))

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
	for _, track := range negotidated {
		switch track.MediaType() {
		case types.MediaTypeVideo:
			videoCodec, err := track.VideoCodec()
			if err != nil {
				return err
			}

			trackID, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			webrtcCodecCapability, err := videoCodec.WebRTCCodecCapability()
			if err != nil {
				return err
			}
			localTrack, err := pion.NewTrackLocalStaticRTP(webrtcCodecCapability, trackID.String(), streamID.String())
			if err != nil {
				return err
			}
			if _, err := pc.AddTrack(localTrack); err != nil {
				return err
			}

			h.localTracks[track] = localTrack
		case types.MediaTypeAudio:
			audioCodec, err := track.AudioCodec()
			if err != nil {
				return err
			}
			trackID, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			webrtcCodecCApability, err := audioCodec.WebRTCCodecCapability()
			if err != nil {
				return err
			}
			localTrack, err := pion.NewTrackLocalStaticRTP(webrtcCodecCApability, trackID.String(), streamID.String())
			if err != nil {
				return err
			}
			if _, err := pc.AddTrack(localTrack); err != nil {
				return err
			}
			h.localTracks[track] = localTrack
		}
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
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
		utils.SendOrDrop(onConnectionState, connectionState)
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

	h.api = api
	h.pc = pc
	h.negotidated = negotidated
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	h.pc.Close()
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track *hubs.Track) (*TrackContext, error) {
	localTrack, ok := h.localTracks[track]
	if !ok {
		return nil, errors.New("handl not found")
	}
	index := slices.Index(h.negotidated, track)
	sender := h.pc.GetTransceivers()[index].Sender()
	go func() {
		if err := h.HandlerRTCP(ctx, sender); err != nil {
			log.Logger.Error("failed to handl RTCP", zap.Error(err))
		}
	}()
	go func() {
		if err := h.handleSendSenderReport(ctx, sender); err != nil {
			log.Logger.Error("failed to handl send sender report", zap.Error(err))
		}
	}()

	ssrc := uint32(sender.GetParameters().Encodings[0].SSRC)
	pt := uint8(sender.GetParameters().Codecs[0].PayloadType)
	clockRate := sender.GetParameters().Codecs[0].ClockRate

	var packetizer rtp.Packetizer
	switch types.CodecTypeFromMimeType(sender.GetParameters().Codecs[0].MimeType) {
	case types.CodecTypeH264:
		packetizer = rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.H264Payloader{}, rtp.NewRandomSequencer(), clockRate)
	case types.CodecTypeVP8:
		packetizer = rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.VP8Payloader{}, rtp.NewRandomSequencer(), clockRate)
	case types.CodecTypeOpus:
		packetizer = rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.OpusPayloader{}, rtp.NewRandomSequencer(), clockRate)
	default:
		return nil, errors.New("unknown codec type")
	}

	return &TrackContext{
		sourceTrack: track,
		localTrack:  localTrack,
		sender:      sender,
		packetizer:  packetizer,
		buf:         make([]byte, types.ReadBufferSize),
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	packetizer := trackCtx.packetizer
	buf := trackCtx.buf
	localTrack := trackCtx.localTrack
	track := trackCtx.sourceTrack
	if track.CodecType() == types.CodecTypeH264 {
		if h264.NALUType(unit.Payload[0]&0x1f) == h264.NALUTypeIDR {
			codec, _ := track.VideoCodec()
			h264Codec := codec.(*hubcodecs.H264)
			_ = packetizer.Packetize(h264Codec.SPS(), 3000)
			_ = packetizer.Packetize(h264Codec.PPS(), 3000)
		}
	}
	for _, rtpPacket := range packetizer.Packetize(unit.Payload, 3000) { //todo 추상화 필요. h264로 가정함.
		n, err := rtpPacket.MarshalTo(buf)
		if err != nil {
			fmt.Println("marshal rtp err:", err)
			continue
		}

		if _, err := localTrack.Write(buf[:n]); err != nil {
			fmt.Println("write rtp err:", err)
		}
		trackCtx.stats.sendCount.Add(1)
		trackCtx.stats.sendLength.Add(uint32(n))
		trackCtx.stats.lastTS.Store(rtpPacket.Timestamp)
	}

	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	packetizer := trackCtx.packetizer
	buf := trackCtx.buf
	localTrack := trackCtx.localTrack
	for _, rtpPacket := range packetizer.Packetize(unit.Payload, 960) { // todo. 추상화 필요. opus 로 가정함
		n, err := rtpPacket.MarshalTo(buf)
		if err != nil {
			fmt.Println("marshal rtp err:", err)
			continue
		}

		if _, err := localTrack.Write(buf[:n]); err != nil {
			fmt.Println("write rtp err:", err)
		}
		trackCtx.stats.sendCount.Add(1)
		trackCtx.stats.sendLength.Add(uint32(n))
		trackCtx.stats.lastTS.Store(rtpPacket.Timestamp)
	}
	return nil
}

func (h *Handler) HandlerRTCP(ctx context.Context, sender *pion.RTPSender) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		rtcpPackets, _, err := sender.ReadRTCP()
		if err != nil {
			return err
		}
		for _, rtcpPacket := range rtcpPackets {
			_ = rtcpPacket
			// TODO RTCP 처리
		}
	}
}

func (h *Handler) handleSendSenderReport(ctx context.Context, sender *pion.RTPSender) error {
	var lastTS atomic.Uint32
	var sendCount atomic.Uint32
	var sendLength atomic.Uint32

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			sr := rtcp.SenderReport{
				SSRC:        uint32(sender.GetParameters().Encodings[0].SSRC),
				NTPTime:     uint64(ntp.GetNTPTime(time.Now())),
				RTPTime:     lastTS.Load(),
				PacketCount: sendCount.Load(),
				OctetCount:  sendLength.Load(),
			}
			if err := h.pc.WriteRTCP([]rtcp.Packet{&sr}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return nil
			}
		}
	}
}
