package writers

import (
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

func NewWriter(index int, timebase int, codecType types.CodecType) *Writer {
	return &Writer{
		index:     index,
		timebase:  timebase,
		filter:    NewFilter(codecType),
		bitStream: NewBitStream(codecType),
	}
}

type Writer struct {
	basePTS int64
	baseDTS int64

	timebase  int
	index     int
	filter    Filter
	bitStream BitStream

	sps, pps []byte // for video (h264)
}

func (v *Writer) WriteAudioPkt(unit units.Unit, pkt *avcodec.Packet) *avcodec.Packet {
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

func (v *Writer) WriteVideoPkt(unit units.Unit, pkt *avcodec.Packet) *avcodec.Packet {
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

	newPTS := avutil.AvRescaleQRound(pts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX)
	pkt.SetPTS(newPTS)
	pkt.SetDTS(avutil.AvRescaleQRound(dts, inputTimebase, outputTimebase, avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, inputTimebase, outputTimebase))
	pkt.SetStreamIndex(v.index)

	data := v.bitStream.SetBitStream(unit)

	pkt.SetData(data)
	pkt.SetFlag(flag)
	return pkt
}
