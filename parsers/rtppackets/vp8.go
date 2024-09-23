package rtppackets

import (
	"github.com/pion/rtp"
	pioncodec "github.com/pion/rtp/codecs"
	"go.uber.org/zap"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/log"
	"sync/atomic"
)

type VP8Parser struct {
	config    *codecs.VP8Config
	fragments []byte

	prevCodec codecs.Codec
	onCodec   atomic.Pointer[func(codec codecs.Codec)]
}

func NewVP8Parser(config *codecs.VP8Config) *VP8Parser {
	return &VP8Parser{
		config: config,
	}
}

func (v *VP8Parser) invokeCodec(codec codecs.Codec) {
	if fn := v.onCodec.Load(); fn != nil {
		(*fn)(codec)
	}
}

func (v *VP8Parser) OnCodec(cb func(codec codecs.Codec)) {
	if cb == nil {
		v.onCodec.Store(nil)
		return
	}
	v.onCodec.Store(&cb)
}

func (v *VP8Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	vp8Payload, err := v.parse(rtpPacket)
	if err != nil {
		log.Logger.Error("vp8 unmarshal err", zap.Error(err))
		return nil
	}
	if vp8Payload == nil {
		return nil
	}

	if err := v.config.Unmarshal(rtpPacket.Payload, vp8Payload); err != nil {
		log.Logger.Error("vp8 config err", zap.Error(err))
	}

	if codec, _ := v.config.GetCodec(); codec != nil {
		if v.prevCodec != codec {
			v.invokeCodec(codec)
			v.prevCodec = codec
		}
	}

	return [][]byte{vp8Payload}
}

func (v *VP8Parser) parse(rtpPacket *rtp.Packet) ([]byte, error) {
	vp8Packet := &pioncodec.VP8Packet{}
	vp8Payload, err := vp8Packet.Unmarshal(rtpPacket.Payload)
	if err != nil {
		return nil, err
	}

	v.fragments = append(v.fragments, vp8Payload...)
	if rtpPacket.Marker {
		fragments := v.fragments
		v.fragments = nil
		return fragments, nil
	}
	return nil, nil
}
