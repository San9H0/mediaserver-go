package vp8

import (
	"bytes"
	"golang.org/x/image/vp8"
	"mediaserver-go/hubs/codecs"
)

type VP8Config struct {
	Width  int
	Height int

	codec codecs.Codec
}

func NewVP8Config() *VP8Config {
	return &VP8Config{}
}

func (v *VP8Config) Unmarshal(rtpPacket, vp8Payload []byte) error {
	s := (rtpPacket[0] & 0x10) >> 4
	if !(vp8Payload[0]&0x01 == 0 && s == 1) { // keyframe
		return nil
	}

	vp8Decoder := vp8.NewDecoder()
	vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
	vp8Frame, err := vp8Decoder.DecodeFrameHeader()
	if err != nil {
		return err
	}

	if v.Width != vp8Frame.Width || v.Height != vp8Frame.Height {
		v.codec, _ = NewVP8(vp8Frame.Width, vp8Frame.Height)
	}
	v.Width = vp8Frame.Width
	v.Height = vp8Frame.Height
	return nil
}

func (v *VP8Config) GetCodec() (codecs.Codec, error) {
	return v.codec, nil
}
