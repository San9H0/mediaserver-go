package packetizers

type H264 struct {
}

func NewH264() *H264 {
	return &H264{}
}

func (h *H264) Parse(payload []byte) ([][]byte, error) {
	return nil, nil
}
