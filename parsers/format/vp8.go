package format

import (
	"bytes"
	"golang.org/x/image/vp8"
)

type VP8Config struct {
	Width  int
	Height int
}

func (v *VP8Config) Unmarshal(rtpPacket, vp8Payload []byte) error {
	vp8Decoder := vp8.NewDecoder()
	vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
	vp8Frame, err := vp8Decoder.DecodeFrameHeader()
	if err != nil {
		return err
	}
	v.Width = vp8Frame.Width
	v.Height = vp8Frame.Height
	return nil
}

func IsVP8KeyFrameSimple(rtpPayload []byte, vp8Payload []byte) bool {
	if len(vp8Payload) == 0 || len(rtpPayload) == 0 {
		return false
	}
	s := (rtpPayload[0] & 0x10) >> 4
	return vp8Payload[0]&0x01 == 0 && s == 1
}

//func MyFunc() {
//	vp8Decoder := vp8.NewDecoder()
//	vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
//	if vp8Frame, err := vp8Decoder.DecodeFrameHeader(); err == nil {
//		if v.codec == nil || (vp8Frame.Width != v.codec.Width() || vp8Frame.Height != v.codec.Height()) {
//			v.codec, _ = codecs.NewVP8(vp8Frame.Width, vp8Frame.Height)
//		}
//	}
//}
