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

	codec codecs.Codec
	cb    func(codec codecs.Codec)
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

	v.fragments = append(v.fragments, vp8Payload...)
	if rtpPacket.Marker {
		fragments := v.fragments
		v.fragments = nil

		keyFrame := fragments[0]&0x01 == 0
		//version := (fragments[0] & 0x0e) >> 1
		//showFrame := (vp8Payload[0] & 0x10) >> 4
		if keyFrame { // keyframe
			config := NewConfig()
			if err := config.Unmarshal(fragments); err == nil {
				vp8Codec := NewVP8(config)
				if v.codec == nil || !vp8Codec.Equals(v.codec) {
					v.codec = vp8Codec
					v.cb(vp8Codec)
				}
			}
		}
		return [][]byte{fragments}
	}
	return nil
}
