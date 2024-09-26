package containers

import (
	"errors"

	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
	"mediaserver-go/utils/types"
)

type Container interface {
	Extension() string
	SetWriteHeader(ctx *avformat.FormatContext) error
	Codecs() []codecs.Codec
}

func NewContainer(video codecs.VideoCodec, audio codecs.AudioCodec) (Container, error) {
	var container Container
	if video != nil {
		switch video.CodecType() {
		case types.CodecTypeVP8:
			container = NewWebM(video, audio)
		case types.CodecTypeH264:
			container = NewMP4(video, audio)
		default:
			return nil, errors.New("unsupported video codec")
		}
	} else if audio != nil {
		switch audio.CodecType() {
		case types.CodecTypeOpus:
			container = NewWebM(video, audio)
		case types.CodecTypeAAC:
			container = NewMP4(video, audio)
		default:
			return nil, errors.New("unsupported audio codec")
		}
	}
	if container == nil {
		return nil, errors.New("unsupported codec")
	}
	return container, nil
}
