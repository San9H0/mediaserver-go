package whep

import (
	"context"
	"encoding/binary"
	"fmt"
	commonh264 "github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"mediaserver-go/codecs/h264"
	"mediaserver-go/egress/sessions/whep/playoutdelay"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/ntp"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
	"time"
)

type RemoteTrackHandler struct {
	mu sync.RWMutex

	packetCh chan Packet
	buf      []byte

	mediaType              types.MediaType
	pc                     *pion.PeerConnection
	localTrack             *pion.TrackLocalStaticRTP
	transceiver            *pion.RTPTransceiver
	sender                 *pion.RTPSender
	stats                  *Stats
	packetizer             rtp.Packetizer
	playoutDelayHandler    *playoutdelay.Handler
	getExtensions          []func() (int, []byte, bool)
	remoteRTXHandler       *remoteRTX
	adaptiveBitrateHandler *ABSHandler
}

type Args struct {
	mediaType              types.MediaType
	localTrack             *pion.TrackLocalStaticRTP
	transceiver            *pion.RTPTransceiver
	sender                 *pion.RTPSender
	stats                  *Stats
	packetizer             rtp.Packetizer
	playoutDelayHandler    *playoutdelay.Handler
	getExtensions          []func() (int, []byte, bool)
	remoteRTXHandler       *remoteRTX
	adaptiveBitrateHandler *ABSHandler
	pc                     *pion.PeerConnection
}

func NewRemoteTrackHandler(args Args) *RemoteTrackHandler {
	return &RemoteTrackHandler{
		mediaType:              args.mediaType,
		packetCh:               make(chan Packet, 100),
		buf:                    make([]byte, types.ReadBufferSize),
		localTrack:             args.localTrack,
		transceiver:            args.transceiver,
		sender:                 args.sender,
		stats:                  args.stats,
		packetizer:             args.packetizer,
		playoutDelayHandler:    args.playoutDelayHandler,
		getExtensions:          args.getExtensions,
		remoteRTXHandler:       args.remoteRTXHandler,
		adaptiveBitrateHandler: args.adaptiveBitrateHandler,
		pc:                     args.pc,
	}
}

type Packet struct {
	unit  units.Unit
	track hubs.Track
	rid   string
}

func (r *RemoteTrackHandler) Run(ctx context.Context) error {
	go r.HandlerRTCP(ctx)
	go r.handleSendSenderReport(ctx)
	if r.mediaType == types.MediaTypeVideo {
		go r.adaptiveBitrateHandler.Run(ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-r.packetCh:
			codec := packet.track.GetCodec()
			var err error
			if codec.MediaType() == types.MediaTypeVideo {
				err = r.onVideo(ctx, packet.track, packet.unit, packet.rid)
			} else {
				err = r.onAudio(ctx, packet.track, packet.unit)
			}
			if err != nil {
				log.Logger.Error("on packet err", zap.Error(err))
			}

		}
	}
}

func (r *RemoteTrackHandler) onVideo(ctx context.Context, track hubs.Track, unit units.Unit, rid string) error {
	if !r.adaptiveBitrateHandler.isCurrentSpatialLayer(rid) {
		return nil
	}
	if !r.adaptiveBitrateHandler.CanSendTemporalLayer(track, unit) {
		r.packetizer.SkipSamples(3000)
		return nil
	}

	if track.GetCodec().CodecType() == types.CodecTypeH264 { //todo 추상화 필요. h264로 가정함.
		if commonh264.NALUType(unit.Payload[0]&0x1f) == commonh264.NALUTypeIDR {
			h264Codec := track.GetCodec().(*h264.H264)
			_ = r.packetizer.Packetize(h264Codec.SPS(), 3000)
			_ = r.packetizer.Packetize(h264Codec.PPS(), 3000)
		}
	}

	rtpPackets := r.packetizer.Packetize(unit.Payload, 3000)
	for _, rtpPacket := range rtpPackets {
		for _, getExt := range r.getExtensions {
			id, payload, ok := getExt()
			if !ok {
				continue
			}
			rtpPacket.Header.SetExtension(uint8(id), payload)
		}

		if err := r.localTrack.WriteRTP(rtpPacket); err != nil {
			fmt.Println("[TESTDEBUG] write err?:", err)
			return err
		}

		r.stats.sendCount.Add(1)
		r.stats.sendLength.Add(uint32(rtpPacket.MarshalSize()))
		r.stats.lastNTP.Store(uint64(ntp.GetNTPTime(time.Now())))
		r.stats.lastTS.Store(rtpPacket.Timestamp)
	}

	return nil
}

func (r *RemoteTrackHandler) onAudio(ctx context.Context, track hubs.Track, unit units.Unit) error {
	for _, rtpPacket := range r.packetizer.Packetize(unit.Payload, 960) { // todo. 추상화 필요. opus 로 가정함
		n, err := rtpPacket.MarshalTo(r.buf)
		if err != nil {
			fmt.Println("marshal rtp err:", err)
			continue
		}

		if _, err := r.localTrack.Write(r.buf[:n]); err != nil {
			return err
		}
		r.stats.sendCount.Add(1)
		r.stats.sendLength.Add(uint32(n))
		r.stats.lastNTP.Store(uint64(ntp.GetNTPTime(time.Now())))
		r.stats.lastTS.Store(rtpPacket.Timestamp)
	}
	return nil
}

func (r *RemoteTrackHandler) HandlerRTCP(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		rtcpPackets, _, err := r.sender.ReadRTCP()
		if err != nil {
			return err
		}
		for _, irtcpPacket := range rtcpPackets {
			switch rtcpPacket := irtcpPacket.(type) {
			case *rtcp.TransportLayerCC:
			case *rtcp.PictureLossIndication:
			case *rtcp.TransportLayerNack:
				r.stats.nackCount.Add(1)
			case *rtcp.ReceiverReport:
			case *rtcp.ReceiverEstimatedMaximumBitrate:
				fmt.Printf("[TESTDEBUG] Rebm packet:%f, %v\n", rtcpPacket.Bitrate, rtcpPacket.SSRCs)
				// TODO
			}
			_ = irtcpPacket
			// TODO RTCP 처리
		}
	}
}

func (r *RemoteTrackHandler) handleSendSenderReport(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			sr := rtcp.SenderReport{
				SSRC:        uint32(r.sender.GetParameters().Encodings[0].SSRC),
				NTPTime:     r.stats.lastNTP.Load(),
				RTPTime:     r.stats.lastTS.Load(),
				PacketCount: r.stats.sendCount.Load(),
				OctetCount:  r.stats.sendLength.Load(),
			}
			if err := r.pc.WriteRTCP([]rtcp.Packet{&sr}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return nil
			}
		}
	}
}

type remoteRTX struct {
	sequenceNumber uint16
	payloadType    uint8
}

func (r *remoteRTX) makeRTXPacket(rtpPacket *rtp.Packet, osn uint16) *rtp.Packet {
	rtxPacket := &rtp.Packet{
		Header:  rtpPacket.Header,
		Payload: make([]byte, 2+len(rtpPacket.Payload)),
	}
	rtxPacket.SequenceNumber = r.getRTXSeq()
	rtxPacket.PayloadType = r.payloadType
	binary.BigEndian.PutUint16(rtxPacket.Payload[:2], osn)
	copy(rtxPacket.Payload[2:], rtpPacket.Payload)
	return rtxPacket
}

func (r *remoteRTX) getRTXSeq() uint16 {
	sn := r.sequenceNumber
	r.sequenceNumber++
	return sn
}
