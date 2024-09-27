package bitstreams

type Empty struct {
}

func (h *Empty) SetBitStream(payload []byte) []byte {
	return payload
}
