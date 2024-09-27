package opus

type Decoder struct{}

func (d *Decoder) KeyFrame(payload []byte) bool {
	return false
}
