package factory

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/aac"
	"mediaserver-go/codecs/av1"
	"mediaserver-go/codecs/h264"
	"mediaserver-go/codecs/opus"
	"mediaserver-go/codecs/vp8"
	"strings"
)

func NewBase(mime string) (codecs.Base, error) {
	switch strings.ToLower(mime) {
	case strings.ToLower(pion.MimeTypeAV1):
		return &av1.Base{}, nil
	case strings.ToLower(pion.MimeTypeVP8):
		return &vp8.Base{}, nil
	case strings.ToLower(pion.MimeTypeH264):
		return &h264.Base{}, nil
	case strings.ToLower(pion.MimeTypeOpus):
		return &opus.Base{}, nil
	case strings.ToLower("audio/aac"):
		return &aac.Base{}, nil
	default:
		return nil, errors.New("unsupported codec")
	}
}
