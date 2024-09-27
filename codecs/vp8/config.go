package vp8

import (
	"bytes"
	"golang.org/x/image/vp8"
)

type Config struct {
	Width  int
	Height int
}

func NewConfig() *Config {
	return &Config{}
}

func (v *Config) Unmarshal(rtpPacket, vp8Payload []byte) error {
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
