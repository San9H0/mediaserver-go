package files

import (
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Filter interface {
	Filter(unit units.Unit) bool
}

func NewFilter(codecType types.CodecType) Filter {
	switch codecType {
	case types.CodecTypeH264:
		return &FilterSPSPPS{}
	default:
		return &FilterEmpty{}
	}
}

type FilterSPSPPS struct {
}

func (f *FilterSPSPPS) Filter(unit units.Unit) bool {
	return h264.NALUType(unit.Payload[0]&0x1f) != h264.NALUTypeSPS && h264.NALUType(unit.Payload[0]&0x1f) != h264.NALUTypePPS
}

type FilterEmpty struct {
}

func (f *FilterEmpty) Filter(unit units.Unit) bool {
	return true
}
