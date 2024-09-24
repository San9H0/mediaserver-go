package bitstreamfilter

import (
	"encoding/binary"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type BitStreamFilter interface {
	AddFilter(unit units.Unit) []byte
	Filter([]byte) [][]byte
}

func NewBitStream(codecType types.CodecType) BitStreamFilter {
	switch codecType {
	case types.CodecTypeH264:
		return &BitStreamAVC{}
	default:
		return &BitStreamEmpty{}
	}
}

// BitStreamAVC is AVC Configuration Format
type BitStreamAVC struct {
}

func (h *BitStreamAVC) AddFilter(unit units.Unit) []byte {
	avc := make([]byte, 4+len(unit.Payload))
	binary.BigEndian.PutUint32(avc, uint32(len(unit.Payload)))
	copy(avc[4:], unit.Payload)
	return avc
}

func (h *BitStreamAVC) Filter(payload []byte) [][]byte {
	if len(payload) < 1 {
		return nil
	}

	const (
		avcSizeLength = 4
	)

	var aus [][]byte
	offset := uint32(0)
	for {
		if offset >= uint32(len(payload))-avcSizeLength {
			break
		}
		auLength := binary.BigEndian.Uint32(payload[offset:])
		offset += avcSizeLength
		au := payload[offset : auLength+offset]
		offset += auLength
		nalUnit := h264.NALUType(au[0] & 0x1F)
		switch nalUnit {
		case h264.NALUTypeSEI, h264.NALUTypeFillerData, h264.NALUTypeAccessUnitDelimiter:
			continue
		}
		aus = append(aus, au)
	}
	return aus
}

type BitStreamAnnexB struct {
}

func (h *BitStreamAnnexB) AddFilter(unit units.Unit) []byte {
	panic("implement me")
}

type BitStreamEmpty struct {
}

func (h *BitStreamEmpty) AddFilter(unit units.Unit) []byte {
	return unit.Payload
}

func (h *BitStreamEmpty) Filter(payload []byte) [][]byte {
	return [][]byte{payload}
}
