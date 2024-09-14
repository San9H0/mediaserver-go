package packetizers

import (
	"bytes"
	"errors"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	rtpcodecs "github.com/pion/rtp/codecs"
	hubcodecs "mediaserver-go/hubs/codecs"
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
		videoCodec, ok := codec.(*hubcodecs.H264)
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
	codec      *hubcodecs.H264
	samples    uint32

	sps, pps []byte
}

func (h *H264Packetizer) Packetize(unit units.Unit) []*rtp.Packet {
	naluTyp := h264.NALUType(unit.Payload[0] & 0x1f)
	if naluTyp == h264.NALUTypeSPS {
		if h.codec == nil || !bytes.Equal(h.codec.SPS(), unit.Payload) {
			h.sps = append(make([]byte, 0, len(unit.Payload)), unit.Payload...)
		}
		return nil
	}
	if naluTyp == h264.NALUTypePPS {
		if h.codec == nil || !bytes.Equal(h.codec.PPS(), unit.Payload) {
			h.pps = append(make([]byte, 0, len(unit.Payload)), unit.Payload...)
		}
		return nil
	}
	if naluTyp == h264.NALUTypeIDR {
		if len(h.sps) > 0 && len(h.pps) > 0 {
			codec, err := hubcodecs.NewH264(h.sps, h.pps)
			if err != nil {
				return nil
			}
			h.codec = codec
			h.sps, h.pps = nil, nil
		}

		_ = h.packetizer.Packetize(h.codec.SPS(), h.samples)
		_ = h.packetizer.Packetize(h.codec.PPS(), h.samples)
	}

	return h.packetizer.Packetize(unit.Payload, h.samples)
}
