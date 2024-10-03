package opus

type BitStreamEmpty struct {
}

func (h *BitStreamEmpty) AddFilter(payload []byte) []byte {
	return payload
}

func (h *BitStreamEmpty) Filter(payload []byte) [][]byte {
	return [][]byte{payload}
}
