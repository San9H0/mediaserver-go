package factory

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/codecs/aac"
	"mediaserver-go/hubs/codecs/av1"
	"mediaserver-go/hubs/codecs/h264"
	"mediaserver-go/hubs/codecs/opus"
	"mediaserver-go/hubs/codecs/vp8"
	"strings"
)

func NewType(mime string) (codecs.CodecType, error) {
	switch strings.ToLower(mime) {
	case strings.ToLower(pion.MimeTypeAV1):
		return &av1.Type{}, nil
	case strings.ToLower(pion.MimeTypeVP8):
		return &vp8.Type{}, nil
	case strings.ToLower(pion.MimeTypeH264):
		return &h264.Type{}, nil
	case strings.ToLower(pion.MimeTypeOpus):
		return &opus.Type{}, nil
	case strings.ToLower("audio/aac"):
		return &aac.Type{}, nil
	default:
		return nil, errors.New("unsupported codec")
	}
}
