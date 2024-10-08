package files

import (
	"context"
	"errors"
	"fmt"
	"mediaserver-go/parsers/bitstreams"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"mediaserver-go/codecs"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
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
	extension  string
	audioStart atomic.Bool // TODO only audio record

	negotiated []hubs.Track

	outputFormatCtx *avformat.FormatContext
}

func NewHandler(path string) *Handler {
	return &Handler{
		path: path,
	}
}

func (h *Handler) NegotiatedTracks() []hubs.Track {
	ret := make([]hubs.Track, 0, len(h.negotiated))
	return append(ret, h.negotiated...)
}

func (h *Handler) Init(ctx context.Context, sources []*hubs.HubSource) error {
	var err error
	var videoCodec codecs.VideoCodec
	var audioCodec codecs.AudioCodec
	var negotiated []hubs.Track
	for _, source := range sources {
		if source.MediaType() == types.MediaTypeVideo {
			if videoCodec, err = source.VideoCodec(); err != nil {
				return fmt.Errorf("video codec not ready: %w", err)
			}
		} else if source.MediaType() == types.MediaTypeAudio {
			if audioCodec, err = source.AudioCodec(); err != nil {
				return fmt.Errorf("audio codec not ready: %w", err)
			}
		}
		codec, _ := source.Codec()
		track := source.GetTrack(codec)
		negotiated = append(negotiated, track)
	}
	extension, err := codecs.GetExtension(videoCodec, audioCodec)
	if err != nil {
		return err
	}

	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", fmt.Sprintf("output.%s", extension)); ret < 0 {
		return errors.New("avformat context allocation failed")
	}

	for _, sourceTrack := range negotiated {
		sourceCodec := sourceTrack.GetCodec()
		outputStream := outputFormatCtx.AvformatNewStream(nil)
		if outputStream == nil {
			return errors.New("avformat stream allocation failed")
		}
		avCodec := avcodec.AvcodecFindEncoder(sourceCodec.AVCodecID())
		if avCodec == nil {
			return errors.New("encoder not found")
		}
		avCodecCtx := avCodec.AvCodecAllocContext3()
		if avCodecCtx == nil {
			return errors.New("codec context allocation failed")
		}
		sourceCodec.SetCodecContext(avCodecCtx, nil)
		if ret := avCodecCtx.AvCodecOpen2(avCodec, nil); ret < 0 {
			return errors.New("codec open failed")
		}
		if ret := avcodec.AvCodecParametersFromContext(outputStream.CodecParameters(), avCodecCtx); ret < 0 {
			return errors.New("codec parameters from context failed")
		}
	}

	outputFormatCtx.SetPb(avformat.AVIOOpenDynBuf())

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
	buf := avformat.AVIOCloseDynBuf(h.outputFormatCtx.Pb())
	h.outputFormatCtx.AvformatFreeContext()

	filepath := fmt.Sprintf("%s.%s", h.path, h.extension)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	n, err := file.Write(buf)
	if err != nil {
		return fmt.Errorf("error copying to file: %w", err)
	}
	log.Logger.Info("file session is finished",
		zap.String("filepath", filepath),
		zap.Int("size", int(n)))
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track hubs.Track) (*TrackContext, error) {
	index := slices.Index(h.negotiated, track)
	outputStream := h.outputFormatCtx.Streams()[index]

	codec := track.GetCodec()
	var bitstream bitstreams.Bitstream
	bitstream = &bitstreams.Empty{}
	if codec.CodecType() == types.CodecTypeH264 {
		bitstream = &bitstreams.AVCC{}
	}
	return &TrackContext{
		codec:        codec,
		outputStream: outputStream,
		writer:       writers.NewWriter(index, outputStream.TimeBase().Den(), track.GetCodec(), codec.Decoder(), bitstream),
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, u units.Unit) error {
	writer := trackCtx.writer
	unit, ok := writer.BitStreamSummary(u)
	if !ok {
		return nil
	}
	pkt := writer.WriteVideoPkt(unit)

	h.audioStart.Store(true)
	h.mu.Lock()
	_ = h.outputFormatCtx.AvInterleavedWriteFrame(pkt)
	h.mu.Unlock()

	pkt.AvPacketUnref()

	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	if !h.audioStart.Load() {
		return nil
	}
	writer := trackCtx.writer
	pkt := writer.WriteAudioPkt(unit)
	if pkt == nil {
		return nil
	}
	h.mu.Lock()
	_ = h.outputFormatCtx.AvInterleavedWriteFrame(pkt)
	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}
