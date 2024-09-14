package sessions

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"io"
	"mediaserver-go/egress/sessions/files"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"os"
	"sync"
)

const (
	bufferSize = 32768
)

type FileSession struct {
	mu sync.RWMutex

	path         string
	extension    string
	ioBuffer     io.ReadWriteSeeker
	tempBuffer   []byte
	sourceTracks []*hubs.Track

	outputFormatCtx *avformat.FormatContext
}

func NewFileSession(path string, sourceTracks []*hubs.Track, ioBuffer io.ReadWriteSeeker) (*FileSession, error) {
	var err error
	var videoCodec codecs.VideoCodec
	var audioCodec codecs.AudioCodec
	for _, sourceTrack := range sourceTracks {
		if sourceTrack.MediaType() == types.MediaTypeVideo {
			if videoCodec, err = sourceTrack.VideoCodec(); err != nil {
				return nil, fmt.Errorf("video codec not ready: %w", err)
			}
		}
		if sourceTrack.MediaType() == types.MediaTypeAudio {
			if audioCodec, err = sourceTrack.AudioCodec(); err != nil {
				return nil, fmt.Errorf("audio codec not ready: %w", err)
			}
		}
	}
	extension, err := getExtension(videoCodec, audioCodec)
	if err != nil {
		return nil, err
	}

	s := &FileSession{
		path:         path,
		extension:    extension,
		ioBuffer:     ioBuffer,
		tempBuffer:   make([]byte, bufferSize),
		sourceTracks: sourceTracks,
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (f *FileSession) init() error {
	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", fmt.Sprintf("output.%s", f.extension)); ret < 0 {
		return errors.New("avformat context allocation failed")
	}
	f.outputFormatCtx = outputFormatCtx

	for _, sourceTrack := range f.sourceTracks {
		sourceCodec, err := sourceTrack.Codec()
		if err != nil {
			return fmt.Errorf("codec not found: %w", err)
		}
		outputStream := outputFormatCtx.AvformatNewStream(nil)
		if outputStream == nil {
			return errors.New("avformat stream allocation failed")
		}
		avCodec := avcodec.AvcodecFindEncoder(types.CodecIDFromType(sourceTrack.CodecType()))
		if avCodec == nil {
			return errors.New("encoder not found")
		}
		avCodecCtx := avCodec.AvCodecAllocContext3()
		if avCodecCtx == nil {
			return errors.New("codec context allocation failed")
		}
		sourceCodec.SetCodecContext(avCodecCtx)
		if ret := avCodecCtx.AvCodecOpen2(avCodec, nil); ret < 0 {
			return errors.New("codec open failed")
		}
		if ret := avcodec.AvCodecParametersFromContext(outputStream.CodecParameters(), avCodecCtx); ret < 0 {
			return errors.New("codec parameters from context failed")
		}
	}

	avioCtx := avformat.AVIoAllocContext(f.outputFormatCtx, f.ioBuffer, &f.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	f.outputFormatCtx.SetPb(avioCtx)

	if f.extension == "mp4" {
		if ret := f.outputFormatCtx.AvformatWriteHeaderWithFMP4("movflags", "frag_keyframe+empty_moov+default_base_moof"); ret < 0 {
			return errors.New("avformat write header failed for fmp4")
		}
	} else {
		if ret := f.outputFormatCtx.AvformatWriteHeader(nil); ret < 0 {
			return errors.New("avformat write header failed")
		}
	}

	return nil
}

func (f *FileSession) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for index, track := range f.sourceTracks {
		g.Go(func() error {
			return f.readTrack(ctx, index, track)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return f.Finish()
}

func (f *FileSession) readTrack(ctx context.Context, index int, track *hubs.Track) error {
	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()
	pkt := avcodec.AvPacketAlloc()
	outputStream := f.outputFormatCtx.Streams()[index]
	writer := files.NewAVPacketWriter(index, outputStream.TimeBase().Den(), track.MediaType(), track.CodecType())

	for {
		select {
		case <-ctx.Done():
			return nil
		case unit, ok := <-consumerCh:
			if !ok {
				return nil
			}

			setPkt := writer.WriteAvPacket(unit, pkt)
			if setPkt == nil {
				continue
			}
			f.mu.Lock()
			_ = f.outputFormatCtx.AvInterleavedWriteFrame(pkt)
			f.mu.Unlock()
			pkt.AvPacketUnref()
		}
	}
}

func (f *FileSession) Finish() error {
	log.Logger.Info("file session finish start")
	f.outputFormatCtx.AvWriteTrailer()
	avformat.AvIoContextFree(f.outputFormatCtx.Pb())
	f.outputFormatCtx.AvformatFreeContext()

	ioBuffer := f.ioBuffer
	f.ioBuffer = nil
	size, err := ioBuffer.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("error seeking to end of buffer: %w", err)
	}
	_, _ = ioBuffer.Seek(0, io.SeekStart)

	filepath := fmt.Sprintf("%s.%s", f.path, f.extension)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, ioBuffer); err != nil {
		return fmt.Errorf("error copying to file: %w", err)
	}
	log.Logger.Info("file session is finished",
		zap.String("filepath", filepath),
		zap.Int("size", int(size)))
	return nil
}

func getExtension(videoCodec codecs.VideoCodec, audioCodec codecs.AudioCodec) (string, error) {
	extension := ""
	if videoCodec != nil && audioCodec != nil {
		switch videoCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "webm"
		case types.CodecTypeH264:
			extension = "mp4"
		default:
			return "", errors.New("unsupported video codec")
		}
	} else if videoCodec != nil {
		switch videoCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "mkv"
		case types.CodecTypeH264:
			extension = "m4v"
		default:
			return "", errors.New("unsupported video codec")
		}
	} else if audioCodec != nil {
		switch audioCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "mka"
		case types.CodecTypeH264:
			extension = "m4a"
		default:
			return "", errors.New("unsupported video codec")
		}
	}
	return extension, nil
}
