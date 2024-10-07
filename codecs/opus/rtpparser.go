package opus

import (
	"github.com/pion/rtp"
	"mediaserver-go/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/units"
	"sync/atomic"
)

type RTPParser struct {
	once atomic.Bool
	cb   func(codec codecs.Codec)
}

func NewRTPParser(cb func(codec codecs.Codec)) *RTPParser {
	return &RTPParser{
		cb: cb,
	}
}

func (r *RTPParser) Parse(rtpPacket *rtp.Packet) ([][]byte, units.FrameInfo) {
	if !r.once.Swap(true) {
		r.cb(NewOpus(NewConfig(Parameters{
			Channels:     2,
			SampleRate:   48000,
			SampleFormat: int(avutil.AV_SAMPLE_FMT_FLT),
		})))
	}
	return [][]byte{rtpPacket.Payload}, units.FrameInfo{}
}
