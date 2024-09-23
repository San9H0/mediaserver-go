package rtppackets

import (
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs"
	"sync"
)

type OpusParser struct {
	config *codecs.OpusConfig
	once   sync.Once
}

func NewOpusParser(config *codecs.OpusConfig) *OpusParser {
	return &OpusParser{
		config: config,
	}
}

func (o *OpusParser) OnCodec(cb func(codec codecs.Codec)) {
	o.once.Do(func() {
		codec, _ := o.config.GetCodec()
		cb(codec)
	})
}

func (o *OpusParser) Parse(rtpPacket *rtp.Packet) [][]byte {
	return [][]byte{rtpPacket.Payload}
}
