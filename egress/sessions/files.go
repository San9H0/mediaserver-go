package sessions

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avformat"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/types"
	"os"
	"sync"
	"time"
)

const (
	bufferSize = 32768
)

type FileSession struct {
	mu sync.RWMutex

	path       string
	ioBuffer   io.ReadWriteSeeker
	tempBuffer []byte
	tracks     []*hubs.Track

	outputFormatCtx        *avformat.FormatContext
	avioCtx                *avformat.AvIOContext
	audioIndex, videoIndex int
}

func NewFileSession(path string, tracks []*hubs.Track, ioBuffer io.ReadWriteSeeker) (FileSession, error) {
	s := FileSession{
		path:       path,
		ioBuffer:   ioBuffer,
		tempBuffer: make([]byte, bufferSize),
		tracks:     tracks,
		audioIndex: -1,
		videoIndex: -1,
	}
	s.init()
	return s, nil
}

func (s *FileSession) init() error {
	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", "output.mp4"); ret < 0 {
		fmt.Println("Error allocating output context", ret)
		return errors.New("avformat context allocation failed")
	}
	s.outputFormatCtx = outputFormatCtx

	for i, track := range s.tracks {
		fmt.Println("[TESTDEBUG] track type:", track.MediaType(), ", codecType:", track.CodecType())
		if track.MediaType() == types.MediaTypeVideo {
			outputVideoStream := outputFormatCtx.AvformatNewStream(nil)
			if outputVideoStream == nil {
				fmt.Println("Error allocating output stream")
				return errors.New("avformat stream allocation failed")
			}

			codec := avcodec.AvcodecFindEncoder(types.CodecIDFromType(track.CodecType()))
			if codec == nil {
				fmt.Println("Error finding encoder")
				return errors.New("encoder not found")
			}

			codecCtx := codec.AvCodecAllocContext3()
			if codecCtx == nil {
				fmt.Println("Error allocating codec context")
				return errors.New("codec context allocation failed")
			}

			var c codecs.VideoCodec
			var err error
			for {
				c, err = track.VideoCodec()
				if err != nil {
					return err
				}
				if c == nil {
					fmt.Println("[TESTDEBUG] video codec wait..")
					time.Sleep(time.Second)
					continue
				}
				break
			}

			codecCtx.SetCodecID(types.CodecIDFromType(c.CodecType()))
			codecCtx.SetCodecType(types.MediaTypeToFFMPEG(c.MediaType()))
			codecCtx.SetWidth(c.Width())
			codecCtx.SetHeight(c.Height())
			codecCtx.SetTimeBase(avutil.NewRational(1, int(c.FPS())))
			codecCtx.SetPixelFormat(avutil.PixelFormat(c.PixelFormat()))
			codecCtx.SetExtraData(c.ExtraData())

			if ret := codecCtx.AvCodecOpen2(codec, nil); ret < 0 {
				fmt.Println("Error opening codec", ret)
				return errors.New("codec open failed")
			}

			if ret := avcodec.AvCodecParametersFromContext(outputVideoStream.CodecParameters(), codecCtx); ret < 0 {
				fmt.Println("Error copying codec parameters from context", ret)
				return errors.New("codec parameters from context failed")
			}
			s.videoIndex = i
		}
		if track.MediaType() == types.MediaTypeAudio {
			outputAudioStream := outputFormatCtx.AvformatNewStream(nil)
			if outputAudioStream == nil {
				fmt.Println("Error allocating output stream")
				return errors.New("avformat stream allocation failed")
			}

			codec := avcodec.AvcodecFindEncoder(types.CodecIDFromType(track.CodecType()))
			if codec == nil {
				fmt.Println("Error finding encoder")
				return errors.New("encoder not found")
			}

			codecCtx := codec.AvCodecAllocContext3()
			if codecCtx == nil {
				fmt.Println("Error allocating codec context")
				return errors.New("codec context allocation failed")
			}

			var c codecs.AudioCodec
			var err error
			for {
				c, err = track.AudioCodec()
				if err != nil {
					return err
				}
				if c == nil {
					fmt.Println("[TESTDEBUG] audio codec wait..")
					time.Sleep(time.Second)
					continue
				}
				break
			}

			codecCtx.SetCodecID(types.CodecIDFromType(c.CodecType()))
			codecCtx.SetCodecType(types.MediaTypeToFFMPEG(c.MediaType()))
			codecCtx.SetSampleRate(c.SampleRate())
			avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), c.Channels())
			codecCtx.SetSampleFmt(avcodec.AvSampleFormat(c.SampleFormat()))

			if ret := codecCtx.AvCodecOpen2(codec, nil); ret < 0 {
				fmt.Println("Error opening codec", ret)
				return errors.New("codec open failed")
			}

			if ret := avcodec.AvCodecParametersFromContext(outputAudioStream.CodecParameters(), codecCtx); ret < 0 {
				fmt.Println("Error copying codec parameters from context", ret)
				return errors.New("codec parameters from context failed")
			}
			s.audioIndex = i
		}
	}

	avioCtx := avformat.AVIoAllocContext(s.outputFormatCtx, s.ioBuffer, &s.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	s.outputFormatCtx.SetPb(avioCtx)
	s.outputFormatCtx.AvformatWriteHeader(nil)
	return nil
}

func (s *FileSession) Run(ctx context.Context) error {
	fmt.Println("[TESTDEBUG] FileSession is run tracks:", len(s.tracks))
	g, ctx := errgroup.WithContext(ctx)
	for _, track := range s.tracks {
		g.Go(func() error {
			return s.readTrack(ctx, track)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	s.Finish()

	return nil
}

func (s *FileSession) readTrack(ctx context.Context, track *hubs.Track) error {
	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()
	pkt := avcodec.AvPacketAlloc()
	_ = pkt
	startKeyFrame := false
	for {
		select {
		case <-ctx.Done():
			return nil
		case unit, ok := <-consumerCh:
			if !ok {
				return nil
			}

			_ = unit
			// Write to output stream
			//fmt.Println("egress kind:", track.MediaType(), ", codecType:", track.CodecType(), ", unit:", unit.PTS)
			if track.MediaType() == types.MediaTypeVideo {
				outputStream := s.outputFormatCtx.Streams()[s.videoIndex]

				if unit.Flags == 1 {
					startKeyFrame = true
				}

				if !startKeyFrame {
					continue
				}

				timebase := avutil.NewRational(1, unit.TimeBase)
				pkt.SetPTS(avutil.AvRescaleQRound(unit.PTS, timebase, outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
				pkt.SetDTS(avutil.AvRescaleQRound(unit.DTS, timebase, outputStream.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX))
				pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, timebase, outputStream.TimeBase()))
				pkt.SetStreamIndex(s.videoIndex)
				//avc := make([]byte, len(unit.Payload))
				//copy(avc, unit.Payload)
				avc := make([]byte, 4+len(unit.Payload))
				binary.BigEndian.PutUint32(avc, uint32(len(unit.Payload)))
				copy(avc[4:], unit.Payload)
				finalLen := 10
				if finalLen > len(avc)-1 {
					finalLen = len(avc)
				}
				pkt.SetData(avc)
				pkt.SetFlag(unit.Flags)
				//fmt.Printf("avc:%X,%X,%X,%X, data:%d\n", data[0], data[1], data[2], data[3], len(data[4:]))

				s.mu.Lock()
				_ = s.outputFormatCtx.AvInterleavedWriteFrame(pkt)
				s.mu.Unlock()
				pkt.AvPacketUnref()
			}

			if track.MediaType() == types.MediaTypeAudio {
				outputStream := s.outputFormatCtx.Streams()[s.audioIndex]

				timebase := avutil.NewRational(1, unit.TimeBase)
				pkt.SetPTS(avutil.AvRescaleQ(unit.PTS, timebase, outputStream.TimeBase()))
				pkt.SetDTS(avutil.AvRescaleQ(unit.DTS, timebase, outputStream.TimeBase()))
				pkt.SetDuration(avutil.AvRescaleQ(unit.Duration, timebase, outputStream.TimeBase()))
				pkt.SetStreamIndex(s.audioIndex)
				pkt.SetData(unit.Payload)
				pkt.SetFlag(unit.Flags)

				s.mu.Lock()
				_ = s.outputFormatCtx.AvInterleavedWriteFrame(pkt)
				s.mu.Unlock()
				pkt.AvPacketUnref()
			}
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

	file, err := os.OpenFile("output5.mp4", os.O_CREATE|os.O_RDWR, 0666)
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
}
