package sessions

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	_ "golang.org/x/image/vp8"

	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/parsers"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type WHIPSession struct {
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState

	stream *hubs.Stream
}

func NewWHIPSession(offer, token string, api *pion.API, stream *hubs.Stream) (WHIPSession, error) {
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return WHIPSession{}, err
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return WHIPSession{}, err
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
		return WHIPSession{}, err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return WHIPSession{}, err
	}

	for range candCh {
	}

	return WHIPSession{
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
		stream:            stream,
	}, nil
}

func (w *WHIPSession) Answer() string {
	return w.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (w *WHIPSession) Run(ctx context.Context) error {
	defer func() {
		w.pc.Close()
		w.stream.Close()
	}()

	var once sync.Once
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			log.Logger.Info("whip ontrack",
				zap.String("mimetype", onTrack.remote.Codec().MimeType),
				zap.String("kind", onTrack.remote.Kind().String()),
				zap.Uint32("ssrc", uint32(onTrack.remote.SSRC())),
			)

			mediaType := types.MediaTypeFromPion(onTrack.remote.Kind())
			codecType := types.CodecTypeFromMimeType(onTrack.remote.Codec().MimeType)

			stats := NewStats(mediaType, onTrack.remote.Codec().ClockRate, uint32(onTrack.remote.SSRC()))

			target := hubs.NewTrack(mediaType, codecType)
			w.stream.AddTrack(target)
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				once.Do(func() {
					go w.sendPLI(ctx, stats)
				})
			} else {
				target.SetCodec(codecs.NewOpus(codecs.OpusParameters{
					SampleRate: int(onTrack.remote.Codec().ClockRate),
					Channels:   int(onTrack.remote.Codec().Channels),
					SampleFmt:  int(avutil.AV_SAMPLE_FMT_S16),
				}))
			}
			go w.sendReceiverReport(ctx, stats)
			go w.readRTP(onTrack.remote, target, stats)
			go w.readRTCP(onTrack.receiver, stats)
		case connectionState := <-w.onConnectionState:
			fmt.Println("conn:", connectionState.String())
			switch connectionState {
			case pion.PeerConnectionStateDisconnected, pion.PeerConnectionStateFailed:
				return nil
			default:
			}
		}
	}
}

func (w *WHIPSession) readRTP(remote *pion.TrackRemote, target *hubs.Track, stats *Stats) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0
	h264Parser := parsers.NewH264Parser()
	vp8Parser := parsers.NewVP8Parser()
	var videoCodec codecs.Codec
	for {
		buf := make([]byte, types.ReadBufferSize)
		n, _, err := remote.Read(buf)
		if err != nil {
			fmt.Println("read rtp err:", err)
			return err
		}
		rtpPacket := &rtp.Packet{}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			fmt.Println("unmarshal rtp err:", err)
			continue
		}

		if startTS == 0 {
			startTS = rtpPacket.Timestamp
		}
		pts := rtpPacket.Timestamp - startTS

		if rtpPacket.Timestamp != prevTS {
			if prevTS == 0 {
				duration = 0
			} else {
				duration = int(rtpPacket.Timestamp - prevTS)
			}
		}

		stats.CalcRTPStats(rtpPacket, n)

		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeH264 {
			aus := h264Parser.Parse(rtpPacket)
			var codec codecs.Codec
			codec = h264Parser.GetCodec()
			if codec == nil {
				continue
			} else if videoCodec != codec {
				target.SetCodec(codec)
				h264codec, ok := codec.(*codecs.H264)
				if !ok {
					log.Logger.Error("codec is not h264")
					continue
				}
				fmt.Printf("profile:%02X, %d, ", h264codec.ExtraData()[1], h264codec.ExtraData()[1])
				fmt.Printf("constraint:%02X, %d, ", h264codec.ExtraData()[2], h264codec.ExtraData()[2])
				fmt.Printf("level:%02X, %d\n", h264codec.ExtraData()[3], h264codec.ExtraData()[3])

				videoCodec = codec
			}

			for _, unit := range aus {
				target.Write(units.Unit{
					Payload:  unit,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: int(remote.Codec().ClockRate),
				})
			}
		}
		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeVP8 {
			aus := vp8Parser.Parse(rtpPacket)
			var codec codecs.Codec
			codec = vp8Parser.GetCodec()
			if codec == nil {
				continue
			} else if videoCodec != codec {
				target.SetCodec(codec)
				videoCodec = codec
			}

			for _, unit := range aus {
				target.Write(units.Unit{
					Payload:  unit,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: int(remote.Codec().ClockRate),
				})
			}
		}
		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeOpus {
			target.Write(units.Unit{
				Payload:  rtpPacket.Payload,
				PTS:      int64(pts),
				DTS:      int64(pts),
				Duration: int64(duration),
				TimeBase: int(remote.Codec().ClockRate),
			})
		}
		prevTS = rtpPacket.Timestamp
	}
}

func (w *WHIPSession) readRTCP(receiver *pion.RTPReceiver, stats *Stats) error {
	for {
		rtcpPackets, _, err := receiver.ReadRTCP()
		if err != nil {
			return err
		}
		for _, irtcpPacket := range rtcpPackets {
			switch rtcpPacket := irtcpPacket.(type) {
			case *rtcp.SenderReport:
				stats.UpdateSR(rtcpPacket)
			default:
			}
		}
	}
}

func (w *WHIPSession) sendReceiverReport(ctx context.Context, stats *Stats) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	prevBytes := uint32(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			totalBytes := stats.getTotalBytes()
			bps := 8 * (totalBytes - prevBytes)
			prevBytes = totalBytes
			_ = bps

			if err := w.pc.WriteRTCP([]rtcp.Packet{stats.getReceiverReport(), stats.getRemB()}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return
			}
		}
	}
}

func (w *WHIPSession) sendPLI(ctx context.Context, stats *Stats) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.pc.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					SenderSSRC: 0, MediaSSRC: stats.ssrc,
				},
			}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return
			}
		}
	}
}

const maxSN = 1 << 16

type Stats struct {
	mu sync.RWMutex

	mediaType types.MediaType
	clockRate uint32
	ssrc      uint32

	// for stats
	totalBytes  uint32
	packetCount uint32
	packetLost  uint32
	maxSeqNo    uint16
	baseSeqNo   uint16
	cycle       uint32

	// for receiver report
	prevExpect     atomic.Uint32
	prevPacketLost atomic.Uint32
	lastSRRTPTime  atomic.Uint32
	lastSRNTPTime  atomic.Uint64
	lastSRTime     atomic.Int64
	jitter         float64
	lastTransit    uint32
}

func NewStats(mediaType types.MediaType, clockRate, ssrc uint32) *Stats {
	return &Stats{
		mediaType: mediaType,
		clockRate: clockRate,
		ssrc:      ssrc,
	}
}

func (s *Stats) CalcRTPStats(pkt *rtp.Packet, n int) {
	arrivalTime := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	sn := pkt.SequenceNumber
	if s.packetCount == 0 {
		s.baseSeqNo = sn
		s.maxSeqNo = sn
	} else if (sn-s.maxSeqNo)&0x8000 == 0 {
		if sn < s.maxSeqNo {
			s.cycle += maxSN
		}
		s.maxSeqNo = sn
	} else if (sn-s.maxSeqNo)&0x8000 > 0 {
		// 재전송필요
	}
	s.packetCount++
	s.totalBytes += uint32(n)

	arrival := uint32(arrivalTime / 1e6 * int64(s.clockRate/1e3))
	transit := arrival - pkt.Timestamp
	if s.lastTransit != 0 {
		d := int32(transit - s.lastTransit)
		if d < 0 {
			d = -d
		}
		s.jitter += (float64(d) - s.jitter) / 16
	}
	s.lastTransit = transit
}

func (s *Stats) UpdateSR(rtcpPacket *rtcp.SenderReport) {
	s.lastSRRTPTime.Store(rtcpPacket.RTPTime)
	s.lastSRNTPTime.Store(rtcpPacket.NTPTime)
	s.lastSRTime.Store(time.Now().UnixNano())
}

func (s *Stats) extMaxSeqNo() uint32 {
	return s.cycle | uint32(s.maxSeqNo)
}

func (s *Stats) getTotalBytes() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalBytes
}

func (s *Stats) getReceiverReport() *rtcp.ReceiverReport {
	s.mu.RLock()
	ssrc := s.ssrc
	jitter := s.jitter
	maxSeqNo := s.maxSeqNo
	packetExpect := s.extMaxSeqNo() - uint32(s.baseSeqNo) + 1
	packetLost := packetExpect - s.packetCount
	s.mu.RUnlock()

	expectedInterval := packetExpect - s.prevExpect.Load()
	lostInterval := packetLost - s.prevPacketLost.Load()
	lostRate := float32(lostInterval) / float32(expectedInterval)
	fractionLost := uint8(lostRate * 256.0)
	lastSenderReport := uint32(s.lastSRNTPTime.Load() >> 16)

	var dlsr uint32
	lastSRtime := s.lastSRTime.Load()
	if lastSRtime != 0 {
		delayMS := uint32((time.Now().Nanosecond() - int(lastSRtime)) / 1e6)
		dlsr = (delayMS / 1e3) << 16
		dlsr |= (delayMS % 1e3) * 65536 / 1000
	}

	s.prevExpect.Store(packetExpect)
	s.prevPacketLost.Store(packetLost)

	return &rtcp.ReceiverReport{
		SSRC: ssrc,
		Reports: []rtcp.ReceptionReport{
			{
				SSRC:               ssrc,
				FractionLost:       fractionLost,
				TotalLost:          packetLost,
				LastSequenceNumber: uint32(maxSeqNo),
				Jitter:             uint32(jitter),
				LastSenderReport:   lastSenderReport,
				Delay:              dlsr,
			},
		},
	}
}

func (s *Stats) getRemB() *rtcp.ReceiverEstimatedMaximumBitrate {
	return &rtcp.ReceiverEstimatedMaximumBitrate{
		Bitrate: float32(3_000_000),
		SSRCs:   []uint32{s.ssrc},
	}
}
