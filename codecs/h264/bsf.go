package h264

import (
	"encoding/binary"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
)

// BitStreamAVC is AVC Configuration Format
type BitStreamAVC struct {
}

func (h *BitStreamAVC) AddFilter(payload []byte) []byte {
	avc := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(avc, uint32(len(payload)))
	copy(avc[4:], payload)
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

func (h *BitStreamAnnexB) AddFilter([]byte) []byte {
	panic("implement me")
}

func (h *BitStreamAnnexB) Filter(payload []byte) [][]byte {
	idx := 0
	unitSize := 0
	var result [][]byte
	for idx < len(payload) {
		if payload[idx] == 0x00 && payload[idx+1] == 0x00 {
			if payload[idx+2] == 0x01 || (payload[idx+2] == 0x00 && payload[idx+3] == 0x01) {
				if unitSize != 0 {
					dst := removeEmulationPreventionBytes(payload[idx-unitSize : unitSize])
					result = append(result, dst)
				}
				if payload[idx+2] == 0x01 {
					idx += 3
				} else {
					idx += 4
				}
			}
		}
		unitSize++
		idx++
	}
	return result
}

func removeEmulationPreventionBytes(src []byte) []byte {
	ret := make([]byte, 0, len(src))
	idx := 0
	copyIdx := 0
	for idx+2 < len(src) {
		if src[idx] == 0x00 && src[idx+1] == 0x00 && src[idx+2] == 0x03 {
			ret = append(ret, src[copyIdx:idx-copyIdx]...)
			ret = append(ret, 0x00, 0x00)
			idx += 3
			copyIdx = idx
			continue
		}
		idx++
	}
	ret = append(ret, src[copyIdx:]...)
	return ret
}
