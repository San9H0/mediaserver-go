package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
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
	"sync"
	"sync/atomic"
	"time"
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
			stats := NewStats(onTrack.remote.Codec().ClockRate, uint32(onTrack.remote.SSRC()))

			mediaType := types.MediaTypeFromPion(onTrack.remote.Kind())
			codecType := types.CodecTypeFromMimeType(onTrack.remote.Codec().MimeType)
			target := hubs.NewTrack(mediaType, codecType)
			w.stream.AddTrack(target)
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				once.Do(func() {
					go w.sendPLI(ctx, stats)
				})
				go w.sendReceiverReport(ctx, stats)
			} else {
				target.SetAudioCodec(codecs.NewOpus(codecs.OpusParameters{
					SampleRate: int(onTrack.remote.Codec().ClockRate),
					Channels:   int(onTrack.remote.Codec().Channels),
					SampleFmt:  int(avutil.AV_SAMPLE_FMT_S16),
				}))
			}
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
	var videoCodec codecs.VideoCodec
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

		stats.CalcRTPStats(rtpPacket)

		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeH264 {
			aus := h264Parser.Parse(rtpPacket)
			var codec codecs.VideoCodec
			codec = h264Parser.GetCodec()
			if codec == nil {
				continue
			} else if videoCodec != codec {
				target.SetVideoCodec(codec)
				videoCodec = codec
			}

			for _, unit := range aus {
				flags := 0
				if h264.NALUType(unit[0]&0x1F) == h264.NALUTypeIDR {
					flags = 1
				}
				target.Write(units.Unit{
					Payload:  unit,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: int(remote.Codec().ClockRate),
					Flags:    flags,
				})
			}
		}
		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeVP8 {
			aus := vp8Parser.Parse(rtpPacket)
			var codec codecs.VideoCodec
			codec = vp8Parser.GetCodec()
			if codec == nil {
				continue
			} else if videoCodec != codec {
				target.SetVideoCodec(codec)
				videoCodec = codec
			}

			for _, unit := range aus {
				flag := 0
				if parsers.IsVP8KeyFrame(unit) {
					flag = 1
				}
				target.Write(units.Unit{
					Payload:  unit,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: int(remote.Codec().ClockRate),
					Flags:    flag,
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
				Flags:    0,
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.pc.WriteRTCP([]rtcp.Packet{stats.getReceiverReport()}); err != nil {
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

	clockRate uint32
	ssrc      uint32

	// for stats
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

func NewStats(clockRate, ssrc uint32) *Stats {
	return &Stats{
		clockRate: clockRate,
		ssrc:      ssrc,
	}
}

func (s *Stats) CalcRTPStats(pkt *rtp.Packet) {
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
