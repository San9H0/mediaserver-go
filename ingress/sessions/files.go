package sessions

import "C"
import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avformat"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/parser/format"
	"mediaserver-go/utils/buffers"
	"mediaserver-go/utils/types"
	"os"
	"slices"
	"time"
)

const (
	bufferSize = 32768
)

type FileSession struct {
	live bool

	audioIndex, videoIndex int

	inputFormatCtx *avformat.FormatContext

	outputFormatCtx   *avformat.FormatContext
	outputVideoStream *avformat.Stream
	outputAudioStream *avformat.Stream

	tempBuffer []byte
	ioBuffer   io.ReadWriteSeeker
}

func NewFileSession(path string, mediaTypes []types.MediaType, live bool) (FileSession, error) {
	audioIndex, videoIndex := -1, -1
	inputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatOpenInput(&inputFormatCtx, path, nil, nil); ret < 0 {
		return FileSession{}, errors.New("avformat open input failed")
	}
	if ret := inputFormatCtx.AvformatFindStreamInfo(nil); ret < 0 {
		return FileSession{}, errors.New("avformat find stream info failed")
	}
	tracks := make(map[int]types.Track)
	for i := 0; i < int(inputFormatCtx.NbStreams()); i++ {
		stream := inputFormatCtx.Streams()[i]
		codecType := types.CodecTypeFromFFMPEG(stream.CodecParameters().CodecID())
		mediaType := types.MediaTypeFromFFMPEG(stream.CodecParameters().CodecType())
		if !slices.Contains(mediaTypes, mediaType) {
			continue
		}

		if mediaType == types.MediaTypeVideo {
			videoIndex = i
		} else if mediaType == types.MediaTypeAudio {
			audioIndex = i
		}
		tracks[i] = types.NewTrack(mediaType, codecType)
	}
	if audioIndex == -1 && videoIndex == -1 {
		return FileSession{}, errors.New("no tracks")
	}

	fmt.Println("videoIndex:", videoIndex, ", audioIndex:", audioIndex)

	inputStream := inputFormatCtx.Streams()[videoIndex]
	codecpar := inputStream.CodecParameters()
	fmt.Println("[TESDTEBUG] input codecpar.CodecID():", codecpar.CodecID())
	fmt.Println("[TESDTEBUG] input codecpar.CodecType():", codecpar.CodecType())
	fmt.Println("[TESDTEBUG] input codecpar.Width():", codecpar.Width())
	fmt.Println("[TESDTEBUG] input codecpar.Height():", codecpar.Height())
	fmt.Println("[TESDTEBUG] input codecpar.ChLayout():", codecpar.ChLayout())
	fmt.Println("[TESDTEBUG] input codecpar.SampleRate():", codecpar.SampleRate())
	fmt.Println("[TESDTEBUG] input codecpar.BitRate():", codecpar.BitRate())

	return FileSession{
		live:       live,
		audioIndex: audioIndex,
		videoIndex: videoIndex,

		inputFormatCtx: inputFormatCtx,
		tempBuffer:     make([]byte, bufferSize),
		ioBuffer:       buffers.NewMemoryBuffer(),
	}, nil
}

func (s *FileSession) Run(ctx context.Context) error {
	count := 0
	defer func() {
		s.inputFormatCtx.AvformatCloseInput()
		fmt.Println("file session done count:", count)
		s.Finish()
		fmt.Println("file write done")
	}()

	fmt.Println("[TESTDEBUG] fileSession run")

	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", "output.mp4"); ret < 0 {
		fmt.Println("Error allocating output context", ret)
		return errors.New("avformat context allocation failed")
	}
	s.outputFormatCtx = outputFormatCtx

	if s.videoIndex != -1 {
		outputVideoStream := outputFormatCtx.AvformatNewStream(nil)
		if outputVideoStream == nil {
			fmt.Println("Error allocating output stream")
			return errors.New("avformat stream allocation failed")
		}
		s.outputVideoStream = outputVideoStream

		codec := avcodec.AvcodecFindEncoder(avcodec.AV_CODEC_ID_H264)
		if codec == nil {
			fmt.Println("Error finding encoder")
			return errors.New("encoder not found")
		}

		codecCtx := codec.AvCodecAllocContext3()
		if codecCtx == nil {
			fmt.Println("Error allocating codec context")
			return errors.New("codec context allocation failed")
		}

		inputStream := s.inputFormatCtx.Streams()[s.videoIndex]

		inputCodecpar := inputStream.CodecParameters()
		sps, pps := format.SPSPPSFromAVCCExtraData(inputCodecpar.ExtraData())
		if len(sps) == 0 || len(pps) == 0 {
			fmt.Println("Error getting sps pps")
			return errors.New("sps pps not found")
		}

		codecCtx.SetCodecID(inputCodecpar.CodecID())
		codecCtx.SetCodecType(inputCodecpar.CodecType())
		codecCtx.SetBitRate(inputCodecpar.BitRate())
		codecCtx.SetWidth(inputCodecpar.Width())
		codecCtx.SetHeight(inputCodecpar.Height())
		codecCtx.SetTimeBase(avutil.NewRational(1, 30))
		codecCtx.SetPixelFormat(avutil.AV_PIX_FMT_YUV420P)
		extradata := format.ExtraDataForAVCC(sps, pps)
		codecCtx.SetExtraData(extradata)

		if ret := codecCtx.AvCodecOpen2(codec, nil); ret < 0 {
			fmt.Println("Error opening codec", ret)
			return errors.New("codec open failed")
		}

		if ret := avcodec.AvCodecParametersFromContext(outputVideoStream.CodecParameters(), codecCtx); ret < 0 {
			fmt.Println("Error copying codec parameters from context", ret)
			return errors.New("codec parameters from context failed")
		}
	}

	if s.audioIndex != -1 {
		outputAudioStream := outputFormatCtx.AvformatNewStream(nil)
		if outputAudioStream == nil {
			fmt.Println("Error allocating output stream")
			return errors.New("avformat stream allocation failed")
		}
		s.outputAudioStream = outputAudioStream

		codec := avcodec.AvcodecFindEncoder(avcodec.AV_CODEC_ID_AAC)
		if codec == nil {
			fmt.Println("Error finding encoder")
			return errors.New("encoder not found")
		}

		codecCtx := codec.AvCodecAllocContext3()
		if codecCtx == nil {
			fmt.Println("Error allocating codec context")
			return errors.New("codec context allocation failed")
		}

		inputStream := s.inputFormatCtx.Streams()[s.audioIndex]
		inputCodecpar := inputStream.CodecParameters()

		codecCtx.SetCodecID(inputCodecpar.CodecID())
		codecCtx.SetCodecType(inputCodecpar.CodecType())
		codecCtx.SetSampleRate(inputCodecpar.SampleRate())
		avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), inputCodecpar.Channels())
		codecCtx.SetBitRate(inputCodecpar.BitRate())
		codecCtx.SetSampleFmt(avcodec.AvSampleFormat(inputCodecpar.Format()))

		if ret := codecCtx.AvCodecOpen2(codec, nil); ret < 0 {
			fmt.Println("Error opening codec", ret)
			return errors.New("codec open failed")
		}

		if ret := avcodec.AvCodecParametersFromContext(outputAudioStream.CodecParameters(), codecCtx); ret < 0 {
			fmt.Println("Error copying codec parameters from context", ret)
			return errors.New("codec parameters from context failed")
		}
	}

	avioCtx := avformat.AVIoAllocContext(s.outputFormatCtx, s.ioBuffer, &s.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	s.outputFormatCtx.SetPb(avioCtx)
	s.outputFormatCtx.AvformatWriteHeader(nil)

	pkt := avcodec.AvPacketAlloc()

	var startTime time.Time

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		ret := s.inputFormatCtx.AvReadFrame(pkt)
		if ret < 0 {
			return io.EOF
		}

		if startTime.IsZero() {
			startTime = time.Now()
		}

		if pkt.StreamIndex() == s.videoIndex {
			itb := s.inputFormatCtx.Streams()[pkt.StreamIndex()].TimeBase()
			ptsTimeSec := avutil.GetTimebaseUSec(itb, pkt.PTS())
			diffus := time.Now().Sub(startTime).Microseconds()
			delay := ptsTimeSec - (diffus)

			istream := s.inputFormatCtx.Streams()[s.videoIndex]
			pkt.SetPTS(avutil.AvRescaleQRound(pkt.PTS(), istream.TimeBase(), s.outputVideoStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
			pkt.SetDTS(avutil.AvRescaleQRound(pkt.DTS(), istream.TimeBase(), s.outputVideoStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
			pkt.SetDuration(avutil.AvRescaleQ(pkt.Duration(), istream.TimeBase(), s.outputVideoStream.TimeBase()))
			pkt.SetPOS(-1)

			_ = s.outputFormatCtx.AvInterleavedWriteFrame(pkt)
			pkt.AvPacketUnref()
			if delay > 0 {
				//	time.Sleep(time.Duration(delay) * time.Microsecond)
			}
		} else if pkt.StreamIndex() == s.audioIndex {
			itb := s.inputFormatCtx.Streams()[pkt.StreamIndex()].TimeBase()
			ptsTimeSec := avutil.GetTimebaseUSec(itb, pkt.PTS())
			diffus := time.Now().Sub(startTime).Microseconds()
			delay := ptsTimeSec - (diffus)

			istream := s.inputFormatCtx.Streams()[s.audioIndex]
			pkt.SetPTS(avutil.AvRescaleQ(pkt.PTS(), istream.TimeBase(), s.outputAudioStream.TimeBase()))
			pkt.SetDTS(avutil.AvRescaleQ(pkt.DTS(), istream.TimeBase(), s.outputAudioStream.TimeBase()))
			pkt.SetDuration(avutil.AvRescaleQ(pkt.Duration(), istream.TimeBase(), s.outputAudioStream.TimeBase()))
			pkt.SetPOS(-1)

			_ = s.outputFormatCtx.AvInterleavedWriteFrame(pkt)
			pkt.AvPacketUnref()
			if delay > 0 {
				//	time.Sleep(time.Duration(delay) * time.Microsecond)
			}
		}

	}
}

func (s *FileSession) readFile(ch chan *avcodec.Packet) {
	for {
		select {
		case pkt := <-ch:
			pkt.AvPacketUnref()
		}
	}
}

func (s *FileSession) Finish() {
	s.outputFormatCtx.AvWriteTrailer()
	avformat.AvIoContextFree(s.outputFormatCtx.Pb())
	s.outputFormatCtx.AvformatFreeContext()

	ioBuffer := s.ioBuffer
	s.ioBuffer = nil
	size, err := ioBuffer.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking to end", err)
		return
	}
	_, _ = ioBuffer.Seek(0, io.SeekStart)
	fmt.Println("write size:", size)

	file, err := os.OpenFile("output.mp4", os.O_CREATE|os.O_RDWR, 0666)
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

func checkNAL(b []byte, avc bool) {
	if len(b) < 4 {
		return
	}

	if avc {
		naluSize := binary.BigEndian.Uint32(b[0:4])
		nalu := b[4] & 0x1F
		fmt.Println("AVCC nalu size", naluSize, "nalu type", nalu)
		return
	}

	if b[0] == 0 && b[1] == 0 && b[2] == 0 && b[3] == 1 {
		fmt.Println("annexb 4byte")
	} else if b[0] == 0 && b[1] == 0 && b[2] == 1 {
		fmt.Println("annexb 3byte")
	}
}
