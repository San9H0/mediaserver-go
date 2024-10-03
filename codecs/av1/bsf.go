package av1

import "github.com/bluenviron/mediacommon/pkg/codecs/av1"

type OBUBitStreamFilter struct {
}

func (h *OBUBitStreamFilter) AddFilter(payload []byte) []byte {
	return payload
}

func (h *OBUBitStreamFilter) Filter(payload []byte) [][]byte {
	var result [][]byte
	offset := 0
	for offset < len(payload) {
		pos := offset
		obuType := payload[offset] >> 3
		_ = obuType
		extFlag := (payload[offset] & 0x4) != 0
		hasSize := (payload[offset] & 0x2) != 0
		offset++
		if extFlag {
			offset++
		}
		obuSize := len(payload) - offset
		if hasSize {
			size, sizeN, err := av1.LEB128Unmarshal(payload[offset:])
			if err != nil {
				return nil
			}
			offset += sizeN
			obuSize = int(size)
		}
		result = append(result, payload[pos:offset+obuSize])
		offset += obuSize
	}

	return result
}
