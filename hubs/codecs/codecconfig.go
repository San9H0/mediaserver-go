package codecs

import (
	"errors"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/utils/types"
)

type Config interface {
	GetCodec() (Codec, error)
}

func NewCodecConfig(codecType types.CodecType) (Config, error) {
	switch codecType {
	case types.CodecTypeVP8:
		return NewVP8Config(), nil
	case types.CodecTypeH264:
		return NewH264Config(), nil
	case types.CodecTypeOpus:
		return NewOpusConfig(OpusParameters{
			Channels:   2,
			SampleRate: 48000,
			SampleFmt:  int(avutil.AV_SAMPLE_FMT_FLT),
		}), nil
	default:
		return nil, errors.New("invalid codec type")
	}
}
