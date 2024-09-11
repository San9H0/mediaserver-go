package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"golang.org/x/sync/errgroup"
	"io"
	"mediaserver-go/hubs"
	hubcodecs "mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/generators"
	"mediaserver-go/utils/ntp"
	"mediaserver-go/utils/types"
	"net"
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
	fmt.Println("[TESTDEBUG] Whep offer:", offer)
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)
	id, err := generators.GenerateID()
	if err != nil {
		return WHEPSession{}, err
	}

	me := &pion.MediaEngine{}
	for _, track := range tracks {
		_ = track
		if videoCodec, _ := track.VideoCodec(); videoCodec != nil {
			//videoCodec.Profile()
			// 42001f
			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: pion.RTPCodecCapability{
					MimeType:     types.MimeTypeFromCodecType(videoCodec.CodecType()),
					ClockRate:    90000,
					Channels:     0,
					SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4D401F"),
					RTCPFeedback: nil,
				},
				PayloadType: 127,
			}, pion.RTPCodecTypeVideo); err != nil {
				return WHEPSession{}, err
			}
		}

		if audioCodec, _ := track.AudioCodec(); audioCodec != nil {
			if err := me.RegisterCodec(pion.RTPCodecParameters{
				RTPCodecCapability: pion.RTPCodecCapability{
					MimeType:     types.MimeTypeFromCodecType(audioCodec.CodecType()),
					ClockRate:    uint32(audioCodec.SampleRate()),
					Channels:     uint16(audioCodec.Channels()),
					SDPFmtpLine:  "a=fmtp:111 minptime=10;maxaveragebitrate=96000;stereo=1;sprop-stereo=1;useinbandfec=1",
					RTCPFeedback: nil,
				},
				PayloadType: 111,
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

	for _, track := range tracks {
		_ = track
		if videoCodec, _ := track.VideoCodec(); videoCodec != nil {
			//_ = videoCodec.Profile()
			localTrack, err := pion.NewTrackLocalStaticRTP(pion.RTPCodecCapability{
				MimeType:     types.MimeTypeFromCodecType(videoCodec.CodecType()),
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4D401F"),
				RTCPFeedback: nil,
			}, "videoTrackID", "streamID")
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

		}
		if audioCodec, _ := track.AudioCodec(); audioCodec != nil {
			localTrack, err := pion.NewTrackLocalStaticRTP(pion.RTPCodecCapability{
				MimeType:     types.MimeTypeFromCodecType(audioCodec.CodecType()),
				ClockRate:    uint32(audioCodec.SampleRate()),
				Channels:     uint16(audioCodec.Channels()),
				SDPFmtpLine:  "a=fmtp:111 minptime=10;maxaveragebitrate=96000;stereo=1;sprop-stereo=1;useinbandfec=1",
				RTCPFeedback: nil,
			}, "audioTrackID", "streamID")
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
		}
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return WHEPSession{}, err
	}

	candCh := make(chan *pion.ICECandidate, 10)
	pc.OnICECandidate(func(candidate *pion.ICECandidate) {
		if candidate == nil {
			close(candCh)
			return
		}
		candCh <- candidate
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

	for range candCh {
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
	fmt.Println("sdp answer:", w.pc.LocalDescription().SDP)
	return w.pc.LocalDescription().SDP
}

func (w *WHEPSession) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
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
	fmt.Println("[TESTDEBUG] Session Started")
	defer func() {
		fmt.Println("[TESTDEBUG] Session Closing")
		w.pc.Close()
		fmt.Println("[TESTDEBUG] Session Closed")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			fmt.Println("whep track:", onTrack.remote.ID())
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
	target, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", 5004))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, target)
	if err != nil {
		return err
	}

	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()
	startKeyFrame := false
	_ = startKeyFrame

	ssrc := uint32(sender.GetParameters().Encodings[0].SSRC)
	pt := uint8(sender.GetParameters().Codecs[0].PayloadType)
	packetizer := rtp.NewPacketizer(types.MTUSize, pt, ssrc, &codecs.H264Payloader{}, rtp.NewRandomSequencer(), 90000)
	_ = packetizer
	fmt.Println("[TESTDEBUG] pt:", pt, ", ssrc:", ssrc)
	lastTS := uint32(0)
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
					RTPTime:     lastTS,
					PacketCount: sendCount.Load(),
					OctetCount:  sendLength.Load(),
				}
				if err := w.pc.WriteRTCP([]rtcp.Packet{&sr}); err != nil {
					fmt.Println("write rtcp err:", err)
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
				if h264.NALUType(unit.Payload[0]&0x1f) == h264.NALUTypeIDR {
					codec, _ := track.VideoCodec()
					h264Codec := codec.(*hubcodecs.H264)
					_ = packetizer.Packetize(h264Codec.SPS(), 3000)
					_ = packetizer.Packetize(h264Codec.PPS(), 3000)
				}
				for _, rtpPacket := range packetizer.Packetize(unit.Payload, 3000) {
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
					conn.Write(buf[:n])
					lastTS = rtpPacket.Timestamp
				}
			}
			if track.MediaType() == types.MediaTypeAudio {

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
