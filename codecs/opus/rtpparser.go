package opus

import (
	"github.com/pion/rtp"
	"mediaserver-go/codecs"
	"sync/atomic"
)

type OpusParser struct {
	once atomic.Bool
	cb   func(codec codecs.Codec)
}

func NewOpusParser(cb func(codec codecs.Codec)) *OpusParser {
	return &OpusParser{
		cb: cb,
	}
}

func (o *OpusParser) Parse(rtpPacket *rtp.Packet) [][]byte {
	if !o.once.Swap(true) {
		o.cb(nil)
	}
	return [][]byte{rtpPacket.Payload}
}
