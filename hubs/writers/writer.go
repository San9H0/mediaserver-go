package writers

import (
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/av1"
	"mediaserver-go/parsers/bitstreams"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync/atomic"
)

func NewWriter(index int, timebase int, codec codecs.Codec, decoder codecs.Decoder, bitstream bitstreams.Bitstream) *Writer {
	return &Writer{
		pkt:       avcodec.AvPacketAlloc(),
		index:     index,
		timebase:  timebase,
		codec:     codec,
		decoder:   decoder,
		bitstream: bitstream,
		buf:       make([]byte, 0, 1024),
	}
}

type Writer struct {
	setPTS  bool
	basePTS int64

	timebase  int
	index     int
	codec     codecs.Codec
	decoder   codecs.Decoder
	bitstream bitstreams.Bitstream
	pkt       *avcodec.Packet

	sps, pps []byte // for video (h264)

	keyFramePTS int64

	start    atomic.Bool
	keyFrame bool

	buf []byte
}

func (v *Writer) WriteAudioPkt(unit units.Unit) *avcodec.Packet {
	if !v.setPTS {
		v.setPTS = true
		v.basePTS = unit.PTS
	}
	pts := unit.PTS - v.basePTS
	dts := unit.DTS - v.basePTS

	inputTimebase := avutil.NewRational(1, unit.TimeBase)
	outputTimebase := avutil.NewRational(1, v.timebase)
	v.pkt.SetPTS(avutil.AvRescaleQ(pts, inputTimebase, outputTimebase))
	v.pkt.SetDTS(avutil.AvRescaleQ(dts, inputTimebase, outputTimebase))
	v.pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
	v.pkt.SetStreamIndex(v.index)
	v.pkt.SetData(unit.Payload)
	v.pkt.SetFlag(0)
	return v.pkt
}

func (v *Writer) BitStreamSummary(unit units.Unit) (units.Unit, bool) {
	if v.codec.CodecType() == types.CodecTypeVP8 {
		keyFrame := unit.Payload[0]&0x01 == 0
		if keyFrame {
			v.keyFrame = true
		}
		v.buf = unit.Payload
	} else if v.codec.CodecType() == types.CodecTypeH264 {
		naluType := h264.NALUType(unit.Payload[0] & 0x1f)
		if naluType == h264.NALUTypeSEI {
			return units.Unit{}, false // drop
		}
		if naluType == h264.NALUTypeSPS || naluType == h264.NALUTypePPS {
			return units.Unit{}, false // drop
		}
		if naluType == h264.NALUTypeIDR {
			v.keyFrame = true
		}
		v.buf = unit.Payload
	} else if v.codec.CodecType() == types.CodecTypeAV1 {
		offset := 0
		obuType := unit.Payload[offset] >> 3
		extFlag := (unit.Payload[offset] & 0x4) != 0
		hasSize := (unit.Payload[0] & 0x2) != 0
		offset++
		if obuType == byte(av1.OBUTypeSequenceHeader) {
			v.keyFrame = true
		}
		if extFlag {
			offset++
		}
		if !hasSize {
			obuSize := len(unit.Payload) - offset
			v.buf = append(v.buf, unit.Payload[0]|0x02) // add hasSize field
			v.buf = append(v.buf, av1.LEB128Marshal(uint(obuSize))...)
			v.buf = append(v.buf, unit.Payload[1:]...)
		} else {
			v.buf = append(v.buf, unit.Payload...)
		}
	}

	if !unit.Marker {
		return units.Unit{}, false
	}

	if v.keyFrame {
		v.start.Store(true)
	}

	if !v.start.Load() {
		v.buf = v.buf[:0]
		return units.Unit{}, false
	}
	keyFrame := 0
	if v.keyFrame {
		keyFrame = 1
	}
	buf := v.buf
	v.buf = v.buf[:0]
	v.keyFrame = false
	return units.Unit{
		Payload:  buf,
		PTS:      unit.PTS,
		DTS:      unit.DTS,
		Duration: unit.Duration,
		TimeBase: unit.TimeBase,
		Marker:   unit.Marker,
		FrameInfo: units.FrameInfo{
			Flag: keyFrame,
		},
	}, true
}

func (v *Writer) WriteVideoPkt(unit units.Unit) *avcodec.Packet {
	inputTimebase := avutil.NewRational(1, unit.TimeBase)
	outputTimebase := avutil.NewRational(1, v.timebase)

	if !v.setPTS {
		v.setPTS = true
		v.basePTS = unit.PTS
	}
	pts := unit.PTS - v.basePTS
	dts := unit.DTS - v.basePTS
	pkt := v.pkt

	pkt.SetPTS(avutil.AvRescaleQRound(pts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDTS(avutil.AvRescaleQRound(dts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
	pkt.SetStreamIndex(v.index)

	pkt.SetFlag(unit.FrameInfo.Flag)

	data := v.bitstream.SetBitStream(unit.Payload)
	pkt.SetData(data)

	return pkt
}
