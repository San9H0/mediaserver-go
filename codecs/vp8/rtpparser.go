package vp8

import (
	"github.com/pion/rtp"
	pioncodec "github.com/pion/rtp/codecs"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
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

func (v *RTPParser) Parse(rtpPacket *rtp.Packet) ([][]byte, units.FrameInfo) {
	vp8Packet := &pioncodec.VP8Packet{}
	vp8Payload, err := vp8Packet.Unmarshal(rtpPacket.Payload)
	if err != nil {
		log.Logger.Error("vp8 unmarshal err", zap.Error(err))
		return nil, units.FrameInfo{}
	}

	//vp8PacketTemp := vp8Packet
	//vp8PacketTemp.Payload = nil
	//if b, err := json.MarshalIndent(vp8PacketTemp, "", "  "); err == nil {
	//	fmt.Println("[TESTDEBUG] ReadRTP ssrc:", rtpPacket.SSRC, ", sn:", rtpPacket.SequenceNumber, ", ts:", rtpPacket.Timestamp, ", VP8Packet:", string(b))
	//}

	v.fragments = append(v.fragments, vp8Payload...)
	if rtpPacket.Marker {
		fragments := v.fragments
		v.fragments = nil

		keyFrame := fragments[0]&0x01 == 0
		flag := 0
		if keyFrame {
			flag = 1
		}

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
		return [][]byte{fragments}, units.FrameInfo{
			Flag:          flag,
			PayloadHeader: vp8Packet,
		}
	}
	return nil, units.FrameInfo{}
}
