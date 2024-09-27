package bitstreams

import (
	"encoding/binary"
)

// AVCC is AVC Configuration Format
type AVCC struct {
}

func (a *AVCC) SetBitStream(payload []byte) []byte {
	avc := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(avc, uint32(len(payload)))
	copy(avc[4:], payload)
	return avc
}
