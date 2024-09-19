package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"go.uber.org/zap"

	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

const (
	bufferSize = 1024 * 128
)

type Handler struct {
	mu sync.RWMutex

	path       string
	ioBuffer   io.ReadWriteSeeker
	tempBuffer []byte

	extension       string
	negotiated      []*hubs.Track
	outputFormatCtx *avformat.FormatContext
}

func NewHandler(path string, buffer io.ReadWriteSeeker) *Handler {
	return &Handler{
		path:       path,
		ioBuffer:   buffer,
		tempBuffer: make([]byte, bufferSize),
	}
}

func (h *Handler) NegotiatedTracks() []*hubs.Track {
	ret := make([]*hubs.Track, 0, len(h.negotiated))
	return append(ret, h.negotiated...)
}

func (h *Handler) Init(ctx context.Context, tracks []*hubs.Track) error {
	var err error
	var videoCodec codecs.VideoCodec
	var audioCodec codecs.AudioCodec
	var negotiated []*hubs.Track
	for _, sourceTrack := range tracks {
		if sourceTrack.MediaType() == types.MediaTypeVideo {
			if videoCodec, err = sourceTrack.VideoCodec(); err != nil {
				return fmt.Errorf("video codec not ready: %w", err)
			}
		}
		if sourceTrack.MediaType() == types.MediaTypeAudio {
			if audioCodec, err = sourceTrack.AudioCodec(); err != nil {
				return fmt.Errorf("audio codec not ready: %w", err)
			}
		}
		negotiated = append(negotiated, sourceTrack)
	}
	extension, err := codecs.GetExtension(videoCodec, audioCodec)
	if err != nil {
		return err
	}

	fmt.Println("[TESTDEBUG] extension:", extension)
	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", fmt.Sprintf("output.%s", extension)); ret < 0 {
		return errors.New("avformat context allocation failed")
	}

	for _, sourceTrack := range negotiated {
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

	avioCtx := avformat.AVIoAllocContext(outputFormatCtx, h.ioBuffer, &h.tempBuffer[0], bufferSize, avformat.AVIO_FLAG_WRITE, true)
	outputFormatCtx.SetPb(avioCtx)

	if extension == "mp4" {
		dict := avutil.DictionaryNull()
		avutil.AvDictSet(&dict, "movflags", "frag_keyframe+empty_moov+default_base_moof", 0)
		if ret := outputFormatCtx.AvformatWriteHeader(&dict); ret < 0 {
			return errors.New("avformat write header failed")
		}
	} else {
		if ret := outputFormatCtx.AvformatWriteHeader(nil); ret < 0 {
			return errors.New("avformat write header failed")
		}
	}

	h.outputFormatCtx = outputFormatCtx
	h.extension = extension
	h.negotiated = negotiated
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	log.Logger.Info("file session finish start")
	h.outputFormatCtx.AvWriteTrailer()
	avformat.AvIoContextFree(h.outputFormatCtx.Pb())
	h.outputFormatCtx.AvformatFreeContext()

	ioBuffer := h.ioBuffer
	h.ioBuffer = nil
	size, err := ioBuffer.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("error seeking to end of buffer: %w", err)
	}
	_, _ = ioBuffer.Seek(0, io.SeekStart)

	filepath := fmt.Sprintf("%s.%s", h.path, h.extension)
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

func (h *Handler) OnTrack(ctx context.Context, track *hubs.Track) (*TrackContext, error) {
	index := slices.Index(h.negotiated, track)
	outputStream := h.outputFormatCtx.Streams()[index]
	return &TrackContext{
		sourceTrack:  track,
		pkt:          avcodec.AvPacketAlloc(),
		outputStream: outputStream,
		writer:       writers.NewWriter(index, outputStream.TimeBase().Den(), track.CodecType()),
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	writer := trackCtx.writer
	pkt := trackCtx.pkt
	setPkt := writer.WriteVideoPkt(unit, pkt)
	if setPkt == nil {
		return nil
	}
	h.mu.Lock()
	_ = h.outputFormatCtx.AvInterleavedWriteFrame(setPkt)
	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	writer := trackCtx.writer
	pkt := trackCtx.pkt
	setPkt := writer.WriteAudioPkt(unit, pkt)
	if setPkt == nil {
		return nil
	}
	h.mu.Lock()
	_ = h.outputFormatCtx.AvInterleavedWriteFrame(setPkt)
	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}
