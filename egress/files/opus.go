package files

import "C"
import (
	"context"
	"errors"
	"fmt"
	"io"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avformat"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/utils/units"
	"os"
	"sync/atomic"
)

const (
	bufferSize = 32768
)

type Opus struct {
	started       atomic.Bool
	sampleRate    int
	tempBuffer    []byte
	prevTimestamp int64

	inputFormatCtx *avformat.FormatContext
	inputStream    *avformat.Stream

	outputFormatCtx *avformat.FormatContext
	outputStream    *avformat.Stream
	avioCtx         *avformat.AvIOContext
	ioBuffer        io.ReadWriteSeeker
}

func NewOpus(sampleRate int, ioBuffer io.ReadWriteSeeker) *Opus {
	return &Opus{
		ioBuffer:   ioBuffer,
		sampleRate: sampleRate,
		tempBuffer: make([]byte, bufferSize),
	}
}

func (o *Opus) Setup(ctx context.Context) error {
	if o.started.Swap(true) {
		return nil
	}

	o.inputFormatCtx = avformat.AvformatAllocContext()
	if o.inputFormatCtx == nil {
		fmt.Println("Error allocating input context")
		return errors.New("avformat context allocation failed")
	}

	o.inputStream = o.inputFormatCtx.AvformatNewStream(nil)
	if o.inputStream == nil {
		fmt.Println("Error allocating input stream")
		return errors.New("avformat stream allocation failed")
	}

	o.inputStream.SetTimeBase(1, o.sampleRate)

	inputCodecParameter := o.inputStream.CodecParameters()
	inputCodecParameter.SetCodecType(avutil.AVMEDIA_TYPE_AUDIO)
	inputCodecParameter.SetCodecID(avcodec.AV_CODEC_ID_OPUS)
	inputCodecParameter.SetSampleRate(o.sampleRate)
	inputCodecParameter.AvChannelLayoutDefault(2)

	outputFormatCtx := avformat.NewAvFormatContextNull()
	result := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", "dummy.mkv")
	if result < 0 {
		fmt.Println("Error allocating output context", result)
		return errors.New("avformat context allocation failed")
	}

	o.outputFormatCtx = outputFormatCtx

	o.outputStream = o.outputFormatCtx.AvformatNewStream(nil)
	o.outputStream.SetCodecParamForMuxer(avcodec.AV_CODEC_ID_OPUS, avutil.AVMEDIA_TYPE_AUDIO, 48000)

	avioCtx := avformat.AVIoAllocContext(o.outputFormatCtx, o.ioBuffer, &o.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	o.avioCtx = avioCtx
	o.outputFormatCtx.SetPb(avioCtx)
	o.outputFormatCtx.AvformatWriteHeader(nil)
	return nil
}

func (o *Opus) WritePacket(unit units.Unit) error {
	duration := unit.Timestamp - o.prevTimestamp
	if o.prevTimestamp == 0 {
		duration = int64(o.sampleRate / 960)
	}
	o.prevTimestamp = unit.Timestamp
	pkt := avcodec.AvPacketAlloc()
	pkt.AvPacketFromByteSlice(unit.Payload)

	pkt.SetPTS(avutil.AvRescaleQRound(unit.Timestamp, o.inputStream.TimeBase(), o.outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDTS(avutil.AvRescaleQRound(unit.Timestamp, o.inputStream.TimeBase(), o.outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDuration(avutil.AvRescaleQ(duration, o.inputStream.TimeBase(), o.outputStream.TimeBase()))
	pkt.SetStreamIndex(0) // todo
	_ = o.outputFormatCtx.AvInterleavedWriteFrame(pkt)

	pkt.AvPacketUnref()
	pkt.AvPacketFree()
	return nil
}

func (o *Opus) Finish() {
	if !o.started.Load() {
		return
	}

	o.outputFormatCtx.AvWriteTrailer()
	//pb := o.avioCtx
	//avformat.AVIOCloseP(&pb)
	avformat.AvIoContextFree(o.avioCtx)
	o.outputFormatCtx.AvformatFreeContext()

	ioBuffer := o.ioBuffer
	o.ioBuffer = nil
	size, err := ioBuffer.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking to end", err)
		return
	}
	_, _ = ioBuffer.Seek(0, io.SeekStart)
	fmt.Println("write size:", size)

	file, err := os.OpenFile("output.mkv", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}
	defer file.Close()
	n, err := io.Copy(file, ioBuffer)
	fmt.Println("opus write size:", n, err)
	if err != nil {
		return
	}
	return

}
