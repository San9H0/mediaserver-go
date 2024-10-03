package packetizers

import (
	"errors"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	rtpcodecs "github.com/pion/rtp/codecs"
	hubcodecs "mediaserver-go/codecs"
	h2642 "mediaserver-go/codecs/h264"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
)

type Packetizer interface {
	Packetize(payload []byte) []*rtp.Packet
}

type CommonPacketizer struct {
	samples    uint32
	packetizer rtp.Packetizer
}

func NewPacketizer(parameters engines.RTPCodecParameters, codec hubcodecs.Codec) (Packetizer, error) {
	ssrc := utils.RandomUint32()
	switch parameters.CodecType {
	case types.CodecTypeVP8:
		return &CommonPacketizer{
			samples:    parameters.ClockRate / 30,
			packetizer: rtp.NewPacketizer(types.MTUSize, parameters.PayloadType, ssrc, &rtpcodecs.VP8Payloader{}, rtp.NewRandomSequencer(), parameters.ClockRate),
		}, nil
	case types.CodecTypeH264:
		videoCodec, ok := codec.(*h2642.H264)
		if !ok {
			return nil, errors.New("invalid codec type")
		}
		return &H264Packetizer{
			packetizer: rtp.NewPacketizer(types.MTUSize, parameters.PayloadType, ssrc, &rtpcodecs.H264Payloader{}, rtp.NewRandomSequencer(), parameters.ClockRate),
			sps:        videoCodec.SPS(),
			pps:        videoCodec.PPS(),
			samples:    parameters.ClockRate / 30,
		}, nil
	case types.CodecTypeOpus:
		return &CommonPacketizer{
			samples:    parameters.ClockRate / 50, // 20 ms
			packetizer: rtp.NewPacketizer(types.MTUSize, parameters.PayloadType, ssrc, &rtpcodecs.OpusPayloader{}, rtp.NewRandomSequencer(), parameters.ClockRate),
		}, nil
	default:
		return nil, errors.New("unsupported codec type")
	}

}

func (p *CommonPacketizer) Packetize(payload []byte) []*rtp.Packet {
	return p.packetizer.Packetize(payload, p.samples)
}

type H264Packetizer struct {
	packetizer rtp.Packetizer
	config     h2642.Config
	codec      *h2642.H264
	samples    uint32

	sps, pps []byte

	sps2, pps2 []byte
}

func (h *H264Packetizer) Packetize(payload []byte) []*rtp.Packet {
	naluTyp := h264.NALUType(payload[0] & 0x1f)
	if naluTyp == h264.NALUTypeSPS {
		h.sps2 = payload
		return nil
	}
	if naluTyp == h264.NALUTypePPS {
		h.pps2 = payload
		return nil
	}
	if naluTyp == h264.NALUTypeIDR {
		if err := h.config.UnmarshalFromSPSPPS(h.sps2, h.pps2); err != nil {
			return nil
		}

		_ = h.packetizer.Packetize(h.sps2, h.samples)
		_ = h.packetizer.Packetize(h.pps2, h.samples)
	}

	return h.packetizer.Packetize(payload, h.samples)
}
