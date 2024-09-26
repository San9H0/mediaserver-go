package packetizers

import (
	"errors"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	rtpcodecs "github.com/pion/rtp/codecs"
	hubcodecs "mediaserver-go/hubs/codecs"
	h2642 "mediaserver-go/hubs/codecs/h264"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Packetizer interface {
	Packetize(unit units.Unit) []*rtp.Packet
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

func (p *CommonPacketizer) Packetize(unit units.Unit) []*rtp.Packet {
	return p.packetizer.Packetize(unit.Payload, p.samples)
}

type H264Packetizer struct {
	packetizer rtp.Packetizer
	config     h2642.H264Config
	codec      *h2642.H264
	samples    uint32

	sps, pps []byte

	sps2, pps2 []byte
	codec2     hubcodecs.Codec
}

func (h *H264Packetizer) Packetize(unit units.Unit) []*rtp.Packet {
	naluTyp := h264.NALUType(unit.Payload[0] & 0x1f)
	if naluTyp == h264.NALUTypeSPS {
		h.sps2 = unit.Payload
		return nil
	}
	if naluTyp == h264.NALUTypePPS {
		h.pps2 = unit.Payload
		return nil
	}
	if naluTyp == h264.NALUTypeIDR {
		if err := h.config.UnmarshalFromSPSPPS(h.sps2, h.pps2); err != nil {
			return nil
		}

		if h.codec2, _ = h.config.GetCodec(); h.codec2 == nil {
			return nil
		}
		_ = h.packetizer.Packetize(h.sps2, h.samples)
		_ = h.packetizer.Packetize(h.pps2, h.samples)
	}

	return h.packetizer.Packetize(unit.Payload, h.samples)
}
