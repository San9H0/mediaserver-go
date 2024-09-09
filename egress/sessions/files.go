package sessions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mediaserver-go/goav/avformat"
	"mediaserver-go/utils/units"
)

const (
	bufferSize = 32768
)

type FileSession struct {
	path       string
	ioBuffer   io.ReadWriteSeeker
	tempBuffer []byte

	outputFormatCtx *avformat.FormatContext
	avioCtx         *avformat.AvIOContext
}

func NewFileSession(path string, inputFormatCtx *avformat.FormatContext, streamIndex int, ioBuffer io.ReadWriteSeeker) (FileSession, error) {
	s := FileSession{
		path:       path,
		ioBuffer:   ioBuffer,
		tempBuffer: make([]byte, bufferSize),
	}
	s.init(inputFormatCtx, streamIndex)
	return s, nil
}

func (s *FileSession) init(inputFormatCtx *avformat.FormatContext, streamIndex int) error {
	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", "output.mp4"); ret < 0 {
		fmt.Println("Error allocating output context", ret)
		return errors.New("avformat context allocation failed")
	}

	outputVideoStream := outputFormatCtx.AvformatNewStream(nil)
	if outputVideoStream == nil {
		fmt.Println("Error allocating output stream")
		return errors.New("avformat stream allocation failed")
	}
	codecpar := outputVideoStream.CodecParameters()

	fmt.Println("[TESTDEBUG] codecpar.CodecID():", codecpar.CodecID())
	fmt.Println("[TESTDEBUG] codecpar.CodecType():", codecpar.CodecType())
	fmt.Println("[TESTDEBUG] codecpar.Width():", codecpar.Width())
	fmt.Println("[TESTDEBUG] codecpar.Height():", codecpar.Height())
	fmt.Println("[TESTDEBUG] codecpar.ChLayout():", codecpar.ChLayout())
	fmt.Println("[TESTDEBUG] codecpar.SampleRate():", codecpar.SampleRate())
	fmt.Println("[TESTDEBUG] codecpar.BitRate():", codecpar.BitRate())

	//printf("Codec Type: %d\n", codecpar->codec_type);
	//printf("Codec ID: %d\n", codecpar->codec_id);
	//printf("Bitrate: %ld\n", codecpar->bit_rate);
	//printf("Width: %d\n", codecpar->width);
	//printf("Height: %d\n", codecpar->height);
	//printf("Channels: %d\n", codecpar->channels);
	//printf("Sample Rate: %d\n", codecpar->sample_rate);

	//inputStream := inputFormatCtx.Streams()[streamIndex]
	//if ret := avcodec.AvCodecParametersCopy(outputVideoStream.CodecParameters(), inputStream.CodecParameters()); ret < 0 {
	//	fmt.Println("Error copying codec parameters", ret)
	//	return errors.New("codec parameters copy failed")
	//}
	//
	//outputVideoStream.SetTimeBase(inputStream.TimeBase().Num(), inputStream.TimeBase().Den())
	//
	//avioCtx := avformat.AVIoAllocContext(outputFormatCtx, s.ioBuffer, &s.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	//s.avioCtx = avioCtx
	//s.outputFormatCtx.SetPb(avioCtx)
	//s.outputFormatCtx.AvformatWriteHeader(nil)
	//
	//s.outputFormatCtx = outputFormatCtx
	return nil
}

func (s *FileSession) Run(ctx context.Context, hubCh chan units.Unit) error {
	for {
		select {
		case <-ctx.Done():
			return nil
			//case unit := <-hubCh:
			//	fmt.Println("[TESTDEBUG] unit.Timestamp:", unit.Timestamp)
		}
	}
}
