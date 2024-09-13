package files

import (
	"encoding/binary"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type BitStream interface {
	SetBitStream(unit units.Unit) []byte
}

func NewBitStream(codecType types.CodecType) BitStream {
	switch codecType {
	case types.CodecTypeH264:
		return &BitStreamAVC{}
	default:
		return &BitStreamEmpty{}
	}
}

type BitStreamAVC struct {
}

func (h *BitStreamAVC) SetBitStream(unit units.Unit) []byte {
	avc := make([]byte, 4+len(unit.Payload))
	binary.BigEndian.PutUint32(avc, uint32(len(unit.Payload)))
	copy(avc[4:], unit.Payload)
	return avc
}

type BitStreamAnnexB struct {
}

func (h *BitStreamAnnexB) SetBitStream(unit units.Unit) []byte {
	panic("implement me")
}

type BitStreamEmpty struct {
}

func (h *BitStreamEmpty) SetBitStream(unit units.Unit) []byte {
	return unit.Payload
}
