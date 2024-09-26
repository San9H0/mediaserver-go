package vp8

import (
	"github.com/pion/rtp"
	pioncodec "github.com/pion/rtp/codecs"
	"go.uber.org/zap"
	"mediaserver-go/hubs/codecs/rtpparsers"
	"mediaserver-go/utils/log"
)

type VP8Parser struct {
	fragments []byte

	cb rtpparsers.Callback
}

func NewVP8Parser(cb rtpparsers.Callback) *VP8Parser {
	return &VP8Parser{
		cb: cb,
	}
}

func (v *VP8Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	vp8Packet := &pioncodec.VP8Packet{}
	vp8Payload, err := vp8Packet.Unmarshal(rtpPacket.Payload)
	if err != nil {
		log.Logger.Error("vp8 unmarshal err", zap.Error(err))
		return nil
	}

	var fragments []byte
	v.fragments = append(v.fragments, vp8Payload...)
	if rtpPacket.Marker {
		fragments = v.fragments
		v.fragments = nil
	}

	//if err := v.config.Unmarshal(rtpPacket.Payload, vp8Payload); err != nil {
	//	log.Logger.Error("vp8 factory err", zap.Error(err))
	//}
	//
	//if codec, _ := v.config.GetCodec(); codec != nil {
	//	if v.prevCodec != codec {
	//		v.invokeCodec(codec)
	//		v.prevCodec = codec
	//	}
	//}
	//
	//return [][]byte{vp8Payload}

	return v.cb([][]byte{fragments})

}
