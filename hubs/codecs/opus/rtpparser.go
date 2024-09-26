package opus

import (
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs/rtpparsers"
	"sync/atomic"
)

type OpusParser struct {
	once atomic.Bool
	cb   rtpparsers.Callback
}

func NewOpusParser(cb rtpparsers.Callback) *OpusParser {
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
