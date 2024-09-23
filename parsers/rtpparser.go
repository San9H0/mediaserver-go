package parsers

import (
	"errors"
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/parsers/rtppackets"
)

var (
	errInvalidCodecType = errors.New("invalid codec type")
)

type RTPParser interface {
	OnCodec(cb func(codec codecs.Codec))
	Parse(rtpPacket *rtp.Packet) [][]byte
}

func NewRTPParser(codecConfig codecs.Config, cb func(codec codecs.Codec)) (RTPParser, error) {
	var parser RTPParser
	switch config := codecConfig.(type) {
	case *codecs.VP8Config:
		parser = rtppackets.NewVP8Parser(config)
	case *codecs.H264Config:
		parser = rtppackets.NewH264Parser(config)
	case *codecs.OpusConfig:
		parser = rtppackets.NewOpusParser(config)
	default:
		return nil, errInvalidCodecType
	}
	parser.OnCodec(cb)
	return parser, nil
}
