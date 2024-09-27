package h264

import "github.com/bluenviron/mediacommon/pkg/codecs/h264"

type Decoder struct {
}

func (d *Decoder) KeyFrame(payload []byte) bool {
	return h264.NALUType(payload[0]&0x1f) == h264.NALUTypeIDR
}
