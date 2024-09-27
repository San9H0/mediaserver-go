package vp8

import (
	"github.com/pion/rtp"
	pioncodec "github.com/pion/rtp/codecs"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/utils/log"
)

type RTPParser struct {
	fragments []byte

	cb func(codec codecs.Codec)
}

func NewRTPParser(cb func(codec codecs.Codec)) *RTPParser {
	return &RTPParser{
		cb: cb,
	}
}

func (v *RTPParser) Parse(rtpPacket *rtp.Packet) [][]byte {
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

	s := (rtpPacket.Payload[0] & 0x10) >> 4
	if vp8Payload[0]&0x01 == 0 && s == 1 { // keyframe
		config := NewConfig()
		if err := config.Unmarshal(rtpPacket.Payload, vp8Payload); err == nil {
			v.cb(NewVP8(config))
		}
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

	return [][]byte{fragments}

}
