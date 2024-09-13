package parsers

import (
	"bytes"
	"github.com/pion/rtp"
	pioncodec "github.com/pion/rtp/codecs"
	"golang.org/x/image/vp8"
	"mediaserver-go/hubs/codecs"
)

type VP8Parser struct {
	codec *codecs.VP8

	fragments []byte
}

func NewVP8Parser() *VP8Parser {
	return &VP8Parser{}
}

func (v *VP8Parser) GetCodec() *codecs.VP8 {
	return v.codec
}

func (v *VP8Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	vp8Packet := &pioncodec.VP8Packet{}
	vp8Payload, err := vp8Packet.Unmarshal(rtpPacket.Payload)
	if err != nil {
		return [][]byte{}
	}

	keyFrame := isVP8KeyFrame(vp8Packet, vp8Payload)
	if keyFrame {
		vp8Decoder := vp8.NewDecoder()
		vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
		if vp8Frame, err := vp8Decoder.DecodeFrameHeader(); err == nil {
			if v.codec == nil || (vp8Frame.Width != v.codec.Width() || vp8Frame.Height != v.codec.Height()) {
				v.codec, _ = codecs.NewVP8(vp8Frame.Width, vp8Frame.Height)
			}
		}
	}

	v.fragments = append(v.fragments, vp8Payload...)
	if rtpPacket.Marker {
		fragments := v.fragments
		v.fragments = []byte{}
		return [][]byte{fragments}
	}
	return nil
}

func IsVP8KeyFrame(vp8Payload []byte) bool {
	vp8Decoder := vp8.NewDecoder()
	vp8Decoder.Init(bytes.NewReader(vp8Payload), len(vp8Payload))
	vp8Frame, err := vp8Decoder.DecodeFrameHeader()
	if err != nil {
		return false
	}
	return vp8Frame.KeyFrame
}

func isVP8KeyFrame(vp8Packet *pioncodec.VP8Packet, payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	return payload[0]&0x01 == 0 && vp8Packet.S == 1
}
