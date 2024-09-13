package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	rtpcodecs "github.com/pion/rtp/codecs"
	"github.com/pion/sdp/v3"
	"golang.org/x/sync/errgroup"
	"mediaserver-go/hubs"
	hubcodecs "mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"net"
	"strings"
)

type CodecInfo struct {
	PayloadType uint8
	ClockRate   uint32
}

var payloadTypeMap = map[types.CodecType]CodecInfo{
	types.CodecTypeH264: {
		PayloadType: 127,
		ClockRate:   90000,
	},
}

type RTPSession struct {
	conn         *net.UDPConn
	sourceTracks []*hubs.Track

	targetAddr string
	targetPort int
	ssrc       uint32
	pt         uint8
	clockRate  uint32
	sd         sdp.SessionDescription
}

type RTPTrack struct {
	ssrc      uint32
	pt        uint8
	clockRate uint32
}

func NewRTPSession(targetAddr string, targetPort int, sourceTracks []*hubs.Track) (*RTPSession, error) {
	target, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetAddr, targetPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, target)
	if err != nil {
		return nil, err
	}

	sd := sdp.SessionDescription{}
	sd.Origin = sdp.Origin{
		Username:       "-",
		SessionID:      0,
		SessionVersion: 0,
		NetworkType:    "IN",
		AddressType:    "IP4",
		UnicastAddress: "127.0.0.1",
	}
	sd.ConnectionInformation = &sdp.ConnectionInformation{
		NetworkType: "IN",
		AddressType: "IP4",
		Address: &sdp.Address{
			Address: "127.0.0.1",
		},
	}
	sd.TimeDescriptions = append(sd.TimeDescriptions, sdp.TimeDescription{
		Timing: sdp.Timing{
			StartTime: 0,
			StopTime:  0,
		},
	})

	for _, sourceTrack := range sourceTracks {
		availableCodecs, ok := payloadTypeMap[sourceTrack.CodecType()]
		if !ok {
			return nil, fmt.Errorf("unsupported codec type: %s", sourceTrack.CodecType())
		}

		videoCodec, err := sourceTrack.VideoCodec()
		if err != nil {
			return nil, err
		}
		capa, err := videoCodec.WebRTCCodecCapability()
		if err != nil {
			return nil, err
		}
		md := &sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media: sourceTrack.MediaType().String(),
				Port: sdp.RangedPort{
					Value: targetPort,
				},
				Protos:  []string{"RTP", "AVP"},
				Formats: []string{fmt.Sprintf("%d", availableCodecs.PayloadType)},
			},
		}
		md.Attributes = append(md.Attributes, sdp.Attribute{
			Key:   "rtpmap",
			Value: fmt.Sprintf("%d %s/%d", availableCodecs.PayloadType, strings.ToUpper(string(sourceTrack.CodecType())), capa.ClockRate),
		})
		if capa.SDPFmtpLine != "" {
			md.Attributes = append(md.Attributes, sdp.Attribute{
				Key:   "fmtp",
				Value: fmt.Sprintf("%d %s", availableCodecs.PayloadType, capa.SDPFmtpLine),
			})
		}
		sd.MediaDescriptions = append(sd.MediaDescriptions, md)
	}

	return &RTPSession{
		targetAddr:   targetAddr,
		targetPort:   targetPort,
		conn:         conn,
		sourceTracks: sourceTracks,
		ssrc:         utils.RandomUint32(),
		sd:           sd,
	}, nil
}

func (r *RTPSession) SSRC() uint32 {
	return r.ssrc
}

func (r *RTPSession) PayloadType() uint8 {
	return r.pt
}

func (r *RTPSession) SDP() string {
	b, err := r.sd.Marshal()
	if err != nil {
		return ""
	}
	fmt.Println(string(b))
	return string(b)
}

func (r *RTPSession) Run(ctx context.Context) error {
	defer r.conn.Close()
	g, ctx := errgroup.WithContext(ctx)
	for _, track := range r.sourceTracks {
		g.Go(func() error {
			return r.readTrack(ctx, track)
		})
	}

	return g.Wait()
}

// 현재는 h264만 지원.
func (r *RTPSession) readTrack(ctx context.Context, track *hubs.Track) error {
	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()

	packetizer := rtp.NewPacketizer(types.MTUSize, r.pt, r.ssrc, &rtpcodecs.H264Payloader{}, rtp.NewRandomSequencer(), r.clockRate)
	buf := make([]byte, types.ReadBufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		case unit, ok := <-consumerCh:
			if !ok {
				return nil
			}
			if h264.NALUType(unit.Payload[0]&0x1f) == h264.NALUTypeIDR {
				codec, _ := track.VideoCodec()
				h264Codec := codec.(*hubcodecs.H264)
				_ = packetizer.Packetize(h264Codec.SPS(), 3000)
				_ = packetizer.Packetize(h264Codec.PPS(), 3000)
			}
			for _, rtpPacket := range packetizer.Packetize(unit.Payload, 3000) {
				n, err := rtpPacket.MarshalTo(buf)
				if err != nil {
					continue
				}
				if _, err := r.conn.Write(buf[:n]); err != nil {
					fmt.Println("error writing to conn:", err)
				}
			}
		}
	}
}
