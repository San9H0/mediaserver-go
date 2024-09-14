package files

import (
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type AVPacketWriter interface {
	WriteAvPacket(unit units.Unit, pkt *avcodec.Packet) *avcodec.Packet
}

func NewAVPacketWriter(index int, timebase int, mediaType types.MediaType, codecType types.CodecType) AVPacketWriter {
	if mediaType == types.MediaTypeAudio {
		return NewAudioAVPacketWriter(index, timebase)
	}
	return NewVideoAVPacketWriter(index, timebase, NewFilter(codecType), NewBitStream(codecType))
}

type VideoAVPacketWriter struct {
	basePTS  int64
	baseDTS  int64
	timebase int
	index    int

	filter    Filter
	bitStream BitStream

	sps, pps []byte
}

func NewVideoAVPacketWriter(index int, timebase int, filter Filter, bitStream BitStream) *VideoAVPacketWriter {
	return &VideoAVPacketWriter{
		index:     index,
		timebase:  timebase,
		filter:    filter,
		bitStream: bitStream,
	}
}

func (v *VideoAVPacketWriter) WriteAvPacket(unit units.Unit, pkt *avcodec.Packet) *avcodec.Packet {
	if v.basePTS == 0 {
		v.basePTS = unit.PTS
	}
	pts := unit.PTS - v.basePTS
	if v.baseDTS == 0 {
		v.baseDTS = unit.DTS
	}
	dts := unit.DTS - v.baseDTS

	if v.filter.Drop(unit) {
		return nil
	}

	flag := 0
	if v.filter.KeyFrame(unit) {
		flag = 1
	}

	inputTimebase := avutil.NewRational(1, unit.TimeBase)
	outputTimebase := avutil.NewRational(1, v.timebase)
	pkt.SetPTS(avutil.AvRescaleQRound(pts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDTS(avutil.AvRescaleQRound(dts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
	pkt.SetStreamIndex(v.index)

	data := v.bitStream.SetBitStream(unit)

	pkt.SetData(data)
	pkt.SetFlag(flag)
	return pkt
}

type AudioAVPacketWriter struct {
	basePTS  int64
	baseDTS  int64
	timebase int
	index    int
}

func NewAudioAVPacketWriter(index int, timebase int) *AudioAVPacketWriter {
	return &AudioAVPacketWriter{
		index:    index,
		timebase: timebase,
	}
}

func (v *AudioAVPacketWriter) WriteAvPacket(unit units.Unit, pkt *avcodec.Packet) *avcodec.Packet {
	if v.basePTS == 0 {
		v.basePTS = unit.PTS
	}
	pts := unit.PTS - v.basePTS
	if v.baseDTS == 0 {
		v.baseDTS = unit.DTS
	}
	dts := unit.DTS - v.baseDTS

	inputTimebase := avutil.NewRational(1, unit.TimeBase)
	outputTimebase := avutil.NewRational(1, v.timebase)
	pkt.SetPTS(avutil.AvRescaleQ(pts, inputTimebase, outputTimebase))
	pkt.SetDTS(avutil.AvRescaleQ(dts, inputTimebase, outputTimebase))
	pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
	pkt.SetStreamIndex(v.index)
	pkt.SetData(unit.Payload)
	pkt.SetFlag(0) // 0
	return pkt
}
