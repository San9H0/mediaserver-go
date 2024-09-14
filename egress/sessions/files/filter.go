package files

import (
	"bytes"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"golang.org/x/image/vp8"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Filter interface {
	Drop(unit units.Unit) bool
	KeyFrame(unit units.Unit) bool
}

func NewFilter(codecType types.CodecType) Filter {
	switch codecType {
	case types.CodecTypeH264:
		return &FilterH264{}
	default:
		return &FilterAudio{}
	}
}

type FilterH264 struct {
}

func (f *FilterH264) Drop(unit units.Unit) bool {
	naluType := h264.NALUType(unit.Payload[0] & 0x1f)
	return naluType != h264.NALUTypeSPS && naluType != h264.NALUTypePPS
}

func (f *FilterH264) KeyFrame(unit units.Unit) bool {
	naluType := h264.NALUType(unit.Payload[0] & 0x1f)
	return naluType == h264.NALUTypeIDR
}

type FilterVP8 struct {
}

func (f *FilterVP8) Drop(unit units.Unit) bool {
	return false
}

func (f *FilterVP8) KeyFrame(unit units.Unit) bool {
	vp8Decoder := vp8.NewDecoder()
	vp8Decoder.Init(bytes.NewReader(unit.Payload), len(unit.Payload))
	vp8FrameHeader, err := vp8Decoder.DecodeFrameHeader()
	if err != nil {
		return false
	}
	return vp8FrameHeader.KeyFrame
}

type FilterAudio struct {
}

func (f *FilterAudio) Drop(unit units.Unit) bool {
	return false
}

func (f *FilterAudio) KeyFrame(unit units.Unit) bool {
	return false
}
