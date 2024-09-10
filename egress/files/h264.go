package files

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"io"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avformat"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/parser/codecparser"
	"mediaserver-go/utils/units"
	"os"
	"sync/atomic"
)

type H264 struct {
	started       atomic.Bool
	ioBuffer      io.ReadWriteSeeker
	sampleRate    int
	duration      int
	tempBuffer    []byte
	prevTimestamp int64
	SPS           []byte
	PPS           []byte
	h264parser    codecparser.H264
	pkt           *avcodec.Packet

	inputFormatCtx *avformat.FormatContext
	inputStream    *avformat.Stream

	outputFormatCtx *avformat.FormatContext
	outputStream    *avformat.Stream
	avioCtx         *avformat.AvIOContext
}

func NewH264(sampleRate int, ioBuffer io.ReadWriteSeeker) *H264 {
	fmt.Println("video h264")
	return &H264{
		ioBuffer:   ioBuffer,
		sampleRate: sampleRate,
		duration:   1000 * 960 / 48000,
		tempBuffer: make([]byte, bufferSize),

		pkt: avcodec.AvPacketAlloc(),
	}
}

func (h *H264) Setup(ctx context.Context) error {
	if h.started.Swap(true) {
		return nil
	}

	h.inputFormatCtx = avformat.AvformatAllocContext()
	if h.inputFormatCtx == nil {
		fmt.Println("Error allocating input context")
		return errors.New("avformat context allocation failed")
	}

	h.inputStream = h.inputFormatCtx.AvformatNewStream(nil)
	if h.inputStream == nil {
		fmt.Println("Error allocating input stream")
		return errors.New("avformat stream allocation failed")
	}

	h.inputStream.SetTimeBase(1, h.sampleRate)

	inputCodecParameter := h.inputStream.CodecParameters()
	inputCodecParameter.SetCodecType(avutil.AVMEDIA_TYPE_VIDEO)
	inputCodecParameter.SetCodecID(avcodec.AV_CODEC_ID_H264)
	inputCodecParameter.SetSampleRate(h.sampleRate)
	inputCodecParameter.SetWidth(1280)
	inputCodecParameter.SetHeight(720)
	inputCodecParameter.SetFormat(8)

	outputFormatCtx := avformat.NewAvFormatContextNull()
	result := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", "dummy.mkv")
	if result < 0 {
		fmt.Println("Error allocating output context", result)
		return errors.New("avformat context allocation failed")
	}

	fmt.Println("[TESTDEBUG] setup h.outputFormatCtx")
	h.outputFormatCtx = outputFormatCtx

	h.outputStream = h.outputFormatCtx.AvformatNewStream(nil)
	h.outputStream.SetCodecParamForMuxer(avcodec.AV_CODEC_ID_H264, avutil.AVMEDIA_TYPE_VIDEO, 90000)
	//h.outputStream.SetTimeBase(1, 30)

	avioCtx := avformat.AVIoAllocContext(h.outputFormatCtx, h.ioBuffer, &h.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	h.avioCtx = avioCtx
	h.outputFormatCtx.SetPb(avioCtx)
	h.outputFormatCtx.AvformatWriteHeader(nil)
	return nil
}

func (h *H264) WritePacket(unit units.Unit) error {
	duration := unit.PTS - h.prevTimestamp
	_ = duration
	if h.prevTimestamp == 0 {
		duration = int64(h.sampleRate / 30)
	}
	h.prevTimestamp = unit.PTS

	//aus, err := h.h264parser.GetAU(rtpPacket)
	//if err != nil {
	//	return err
	//}
	//if len(aus) == 0 {
	//	return nil
	//}
	//if !h.h264parser.Ready() {
	//	return nil
	//}

	pkt := avcodec.AvPacketAlloc()
	pkt.AvPacketFromByteSlice(unit.Payload)

	pkt.SetPTS(avutil.AvRescaleQRound(pkt.PTS(), h.inputStream.TimeBase(), h.outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDTS(avutil.AvRescaleQRound(pkt.DTS(), h.inputStream.TimeBase(), h.outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
	pkt.SetDuration(avutil.AvRescaleQ(pkt.Duration(), h.inputStream.TimeBase(), h.outputStream.TimeBase()))
	pkt.SetPOS(-1)

	_ = h.outputFormatCtx.AvInterleavedWriteFrame(pkt)

	pkt.AvPacketUnref()

	return nil
}

func (h *H264) Finish() {
	fmt.Println("[TESTDEBUG] h264 finish")
	if !h.started.Load() {
		return
	}

	h.outputFormatCtx.AvWriteTrailer()
	//pb := o.avioCtx
	//avformat.AVIOCloseP(&pb)
	avformat.AvIoContextFree(h.avioCtx)
	h.outputFormatCtx.AvformatFreeContext()

	ioBuffer := h.ioBuffer
	h.ioBuffer = nil
	size, err := ioBuffer.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking to end", err)
		return
	}
	_, _ = ioBuffer.Seek(0, io.SeekStart)
	fmt.Println("h264 write size:", size)

	file, err := os.OpenFile("video.mkv", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}
	defer file.Close()
	n, err := io.Copy(file, ioBuffer)
	fmt.Println("write size:", n, err)
	if err != nil {
		return
	}
	return
}
