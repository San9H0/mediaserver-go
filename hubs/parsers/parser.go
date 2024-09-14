package parsers

import (
	"errors"
	"github.com/pion/rtp"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/types"
)

var (
	errInvalidCodecType = errors.New("invalid codec type")
)

type Parser interface {
	Parse(payload *rtp.Packet) [][]byte
	GetCodec() codecs.Codec
}

func NewParser(codecType types.CodecType) (Parser, error) {
	switch codecType {
	case types.CodecTypeVP8:
		return NewVP8Parser(), nil
	case types.CodecTypeH264:
		return NewH264Parser(), nil
	case types.CodecTypeOpus:
		return NewAudioParser(codecs.NewOpus(codecs.OpusParameters{
			SampleRate: int(48000),
			Channels:   int(2),
			SampleFmt:  int(avutil.AV_SAMPLE_FMT_S16),
		})), nil
	default:
		return nil, errInvalidCodecType
	}
}

type AudioParser struct {
	codec codecs.Codec
}

func NewAudioParser(codec codecs.Codec) *AudioParser {
	return &AudioParser{
		codec: codec,
	}
}

func (a *AudioParser) Parse(rtpPacket *rtp.Packet) [][]byte {
	return [][]byte{rtpPacket.Payload}
}

func (a *AudioParser) GetCodec() codecs.Codec {
	return a.codec
}
