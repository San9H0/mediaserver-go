package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"io"
	"mediaserver-go/hubs"
	hubcodecs "mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/generators"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/ntp"
	"mediaserver-go/utils/types"
	"sync/atomic"
	"time"
)

type WHEPSession struct {
	id    string
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState

	localTracks []*pion.TrackLocalStaticRTP
	senders     []*pion.RTPSender
	tracks      []*hubs.Track
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func NewWHEPSession(offer, token string, se pion.SettingEngine, tracks []*hubs.Track) (WHEPSession, error) {
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)
	id, err := generators.GenerateID()
	if err != nil {
		return WHEPSession{}, err
	}

	fmt.Println("[TESTDEBUG] whep tracks:", len(tracks))
	me := &pion.MediaEngine{}
	for _, track := range tracks {
		switch track.MediaType() {
		case types.MediaTypeVideo:
			fmt.Println("[TESTDEBUG] whep video...")
			videoCodec, err := track.VideoCodec()
			if err != nil {
				return WHEPSession{}, err
			}
			webrtcCodecCapability, err := videoCodec.WebRTCCodecCapability()
			if err != nil {
				return WHEPSession{}, err
			}
			fmt.Println("[TESTDEBUG] capability:", webrtcCodecCapability.SDPFmtpLine)

			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: webrtcCodecCapability,
				PayloadType:        127,
			}, pion.RTPCodecTypeVideo); err != nil {
				return WHEPSession{}, err
			}
		case types.MediaTypeAudio:
			fmt.Println("[TESTDEBUG] whep audio...")
			audioCodec, err := track.AudioCodec()
			if err != nil {
				return WHEPSession{}, err
			}
			webrtcCodecCapability, err := audioCodec.WebRTCCodecCapability()
			if err != nil {
				return WHEPSession{}, err
			}
			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: webrtcCodecCapability,
				PayloadType:        111,
			}, pion.RTPCodecTypeAudio); err != nil {
				return WHEPSession{}, err
			}
		}
	}

	api := pion.NewAPI(pion.WithMediaEngine(me), pion.WithSettingEngine(se))

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return WHEPSession{}, err
	}

	var localTracks []*pion.TrackLocalStaticRTP
	var senders []*pion.RTPSender

	streamID, err := uuid.NewRandom()
	if err != nil {
		return WHEPSession{}, err
	}
	for _, track := range tracks {
		switch track.MediaType() {
		case types.MediaTypeVideo:
			videoCodec, err := track.VideoCodec()
			if err != nil {
				return WHEPSession{}, err
			}
			fmt.Println("[TESTDEBUG] whep videoCodec...")
			trackID, err := uuid.NewRandom()
			if err != nil {
				return WHEPSession{}, err
			}

			webrtcCodecCapability, err := videoCodec.WebRTCCodecCapability()
			if err != nil {
				return WHEPSession{}, err
			}
			localTrack, err := pion.NewTrackLocalStaticRTP(webrtcCodecCapability, trackID.String(), streamID.String())
			if err != nil {
				fmt.Println("NewTrackLocalStaticRTP err:", err)
				continue
			}
			sender, err := pc.AddTrack(localTrack)
			if err != nil {
				fmt.Println("Video AddTrack err:", err)
				continue
			}
			localTracks = append(localTracks, localTrack)
			senders = append(senders, sender)
			fmt.Println("[TESTDEBUG] videoCodec...senders:", len(senders))
		case types.MediaTypeAudio:
			audioCodec, err := track.AudioCodec()
			if err != nil {
				return WHEPSession{}, err
			}
			fmt.Println("[TESTDEBUG] whep audioCodec...")
			trackID, err := uuid.NewRandom()
			if err != nil {
				return WHEPSession{}, err
			}
			webrtcCodecCApability, err := audioCodec.WebRTCCodecCapability()
			if err != nil {
				return WHEPSession{}, err
			}
			localTrack, err := pion.NewTrackLocalStaticRTP(webrtcCodecCApability, trackID.String(), streamID.String())
			if err != nil {
				fmt.Println("audio NewTrackLocalStaticRTP err:", err)
				continue
			}
			localTrack.Codec()

			sender, err := pc.AddTrack(localTrack)
			if err != nil {
				fmt.Println("Audio AddTrack err:", err)
				continue
			}
			senders = append(senders, sender)
			localTracks = append(localTracks, localTrack)
			fmt.Println("[TESTDEBUG] audioCodec...senders:", len(senders))
		}

	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return WHEPSession{}, err
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
		utils.SendOrDrop(onTrack, OnTrack{
			remote:   remote,
			receiver: receiver,
		})
	})

	sd, err := pc.CreateAnswer(&pion.AnswerOptions{})
	if err != nil {
		return WHEPSession{}, err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return WHEPSession{}, err
	}

	for range candidateCh {
	}

	return WHEPSession{
		id:                id,
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
		tracks:            tracks,
		localTracks:       localTracks,
		senders:           senders,
	}, nil
}

func (w *WHEPSession) Answer() string {
	return w.pc.LocalDescription().SDP
}

func (w *WHEPSession) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	fmt.Println("[TESTDEBUG] w.tracks:", len(w.tracks), ", w.senders", len(w.senders), ", w.localTracks:", w.localTracks)
	for i, track := range w.tracks {
		sender := w.senders[i]
		localTracks := w.localTracks[i]
		g.Go(func() error {
			return w.readTrack(ctx, track, localTracks, sender)
		})
		g.Go(func() error {
			return w.handleRTCP(ctx, sender)
		})
	}
	g.Go(func() error {
		return w.run(ctx)
	})
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (w *WHEPSession) run(ctx context.Context) error {
	defer func() {
		w.pc.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			_ = onTrack

		case connectionState := <-w.onConnectionState:
			fmt.Println("whep conn:", connectionState.String())
			switch connectionState {
			case pion.PeerConnectionStateDisconnected, pion.PeerConnectionStateFailed:
				return io.EOF
			default:
			}
		}
	}
}

func (w *WHEPSession) readTrack(ctx context.Context, track *hubs.Track, localTrack *pion.TrackLocalStaticRTP, sender *pion.RTPSender) error {
	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()

	ssrc := uint32(sender.GetParameters().Encodings[0].SSRC)
	pt := uint8(sender.GetParameters().Codecs[0].PayloadType)
	clockRate := sender.GetParameters().Codecs[0].ClockRate

	fmt.Println("[TESDTEBUG] whep ssrc:", ssrc, "pt:", pt, "clockRate:", clockRate)
	packetizer := rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.H264Payloader{}, rtp.NewRandomSequencer(), clockRate)
	if track.MediaType() == types.MediaTypeAudio {
		packetizer = rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.OpusPayloader{}, rtp.NewRandomSequencer(), clockRate)
	}
	if track.CodecType() == types.CodecTypeVP8 {
		packetizer = rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.VP8Payloader{}, rtp.NewRandomSequencer(), clockRate)
	}

	var lastTS atomic.Uint32
	var sendCount atomic.Uint32
	var sendLength atomic.Uint32
	buf := make([]byte, types.ReadBufferSize)

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sr := rtcp.SenderReport{
					SSRC:        uint32(sender.GetParameters().Encodings[0].SSRC),
					NTPTime:     uint64(ntp.GetNTPTime(time.Now())),
					RTPTime:     lastTS.Load(),
					PacketCount: sendCount.Load(),
					OctetCount:  sendLength.Load(),
				}
				if err := w.pc.WriteRTCP([]rtcp.Packet{&sr}); err != nil {
					log.Logger.Warn("write rtcp err", zap.Error(err))
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case unit, ok := <-consumerCh:
			if !ok {
				return nil
			}

			if track.MediaType() == types.MediaTypeVideo {
				if track.CodecType() == types.CodecTypeH264 {
					if h264.NALUType(unit.Payload[0]&0x1f) == h264.NALUTypeIDR {
						codec, _ := track.VideoCodec()
						h264Codec := codec.(*hubcodecs.H264)
						_ = packetizer.Packetize(h264Codec.SPS(), 3000)
						_ = packetizer.Packetize(h264Codec.PPS(), 3000)
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
						sendCount.Add(1)
						sendLength.Add(uint32(n))
						lastTS.Store(rtpPacket.Timestamp)
					}
				}
				if track.CodecType() == types.CodecTypeVP8 {
					for _, rtpPacket := range packetizer.Packetize(unit.Payload, 3000) { //todo 추상화 필요. h264로 가정함.
						n, err := rtpPacket.MarshalTo(buf)
						if err != nil {
							fmt.Println("marshal rtp err:", err)
							continue
						}

						if _, err := localTrack.Write(buf[:n]); err != nil {
							fmt.Println("write rtp err:", err)
						}
						sendCount.Add(1)
						sendLength.Add(uint32(n))
						lastTS.Store(rtpPacket.Timestamp)
					}
				}
			}
			if track.MediaType() == types.MediaTypeAudio {
				for _, rtpPacket := range packetizer.Packetize(unit.Payload, 960) { // todo. 추상화 필요. opus 로 가정함
					n, err := rtpPacket.MarshalTo(buf)
					if err != nil {
						fmt.Println("marshal rtp err:", err)
						continue
					}

					if _, err := localTrack.Write(buf[:n]); err != nil {
						fmt.Println("write rtp err:", err)
					}
					sendCount.Add(1)
					sendLength.Add(uint32(n))
					lastTS.Store(rtpPacket.Timestamp)
				}
			}
		}
	}
}

func (w *WHEPSession) handleRTCP(ctx context.Context, sender *pion.RTPSender) error {
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
