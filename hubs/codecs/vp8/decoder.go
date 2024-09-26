package vp8

import (
	"bytes"
	"go.uber.org/zap"
	"golang.org/x/image/vp8"
	"mediaserver-go/utils/log"
)

func GetFrameHeader(vp8Payload []byte) (vp8.FrameHeader, bool) {
	if !(vp8Payload[0]&0x01 == 0) {
		return vp8.FrameHeader{}, false
	}
	vp8Decoder := vp8.NewDecoder()
	vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
	vp8Frame, err := vp8Decoder.DecodeFrameHeader()
	if err != nil {
		log.Logger.Error("vp8 decode frame header err", zap.Error(err))
		return vp8.FrameHeader{}, false
	}

	return vp8Frame, true
}
