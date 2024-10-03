package hls

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mediaserver-go/parsers/bitstreams"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"mediaserver-go/codecs"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Endpoint interface {
	SetPayload(payload []byte, name string)
	AppendMedia(payload []byte, index int, duration float64)
}

type Handler struct {
	mu sync.RWMutex

	endpoint Endpoint

	audioStart      atomic.Bool
	extension       string
	negotiated      []hubs.Track
	outputFormatCtx *avformat.FormatContext
	index           int
}

func NewHandler(endpoint Endpoint) *Handler {
	return &Handler{
		endpoint: endpoint,
	}
}

func (h *Handler) CodecString(mediaType types.MediaType) string {
	for _, negotiated := range h.negotiated {
		codec := negotiated.GetCodec()
		if codec.MediaType() == mediaType {
			return codec.HLSMIME()
		}
	}
	return ""
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
		}
		if source.MediaType() == types.MediaTypeAudio {
			if audioCodec, err = source.AudioCodec(); err != nil {
				return fmt.Errorf("audio codec not ready: %w", err)
			}
		}
		codec, _ := source.Codec()
		track := source.GetTrack(codec)
		negotiated = append(negotiated, track)
	}

	if videoCodec.CodecType() != types.CodecTypeH264 && videoCodec.CodecType() != types.CodecTypeAV1 {
		return errors.New("unsupported video codec")
	}

	if audioCodec.CodecType() != types.CodecTypeAAC && audioCodec.CodecType() != types.CodecTypeOpus {
		return errors.New("unsupported audio codec")
	}

	extension := "m4s"
	outputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatAllocOutputContext2(&outputFormatCtx, nil, "", fmt.Sprintf("output.mp4")); ret < 0 {
		return errors.New("avformat context allocation failed")
	}
	h.outputFormatCtx = outputFormatCtx

	for _, track := range negotiated {
		sourceCodec := track.GetCodec()
		outputStream := h.outputFormatCtx.AvformatNewStream(nil)
		if outputStream == nil {
			return errors.New("avformat stream allocation failed")
		}
		avCodec := avcodec.AvcodecFindEncoder(track.GetCodec().AVCodecID())
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
		if sourceCodec.MediaType() == types.MediaTypeVideo {
			outputStream.SetTimeBase(1, 15360)
		}
		fmt.Println("[TESTDEBUG] outputStream.CodecID:", outputStream.CodecParameters().CodecID())
		fmt.Println("[TESTDEBUG] outputStream.CodecType:", outputStream.CodecParameters().CodecType())
		fmt.Println("[TESTDEBUG] outputStream.Width:", outputStream.CodecParameters().Width())
		fmt.Println("[TESTDEBUG] outputStream.Height:", outputStream.CodecParameters().Height())
		fmt.Println("[TESTDEBUG] outputStream.TimeBase:", outputStream.TimeBase())
		fmt.Println("[TESTDEBUG] outputStream.Level:", outputStream.CodecParameters().Level())
		fmt.Println("[TESTDEBUG] outputStream.CodecParameters().ExtraData:", outputStream.CodecParameters().ExtraData())

	}

	h.outputFormatCtx.SetPb(avformat.AVIOOpenDynBuf())

	dict := avutil.DictionaryNull()
	avutil.AvDictSet(&dict, "movflags", "frag_keyframe+empty_moov+default_base_moof", 0)

	if ret := outputFormatCtx.AvformatWriteHeader(&dict); ret < 0 {
		return errors.New("avformat write header failed")
	}

	buf := avformat.AVIOCloseDynBuf(h.outputFormatCtx.Pb())
	h.endpoint.SetPayload(buf, "init.mp4")

	h.outputFormatCtx.SetPb(avformat.AVIOOpenDynBuf())

	h.extension = extension
	h.negotiated = negotiated
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	log.Logger.Info("hls Handler finish start")

	// TODO 남아있는 것 모두 정리.
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track hubs.Track) (*OnTrackContext, error) {
	log.Logger.Info("hls Handler on track", zap.String("codec", track.GetCodec().String()))
	index := slices.Index(h.negotiated, track)
	stream := h.outputFormatCtx.Streams()[index]

	codec := track.GetCodec()

	var bitstream bitstreams.Bitstream
	bitstream = &bitstreams.Empty{}
	if codec.CodecType() == types.CodecTypeH264 {
		bitstream = &bitstreams.AVCC{}
	}

	return &OnTrackContext{
		track:        track,
		outputStream: stream,
		writer:       writers.NewWriter(index, stream.TimeBase().Den(), track.GetCodec(), codec.Decoder(), bitstream),
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *OnTrackContext, u units.Unit) error {
	writer := trackCtx.writer
	unit, ok := writer.BitStreamSummary(u)
	if !ok {
		return nil
	}

	pkt := writer.WriteVideoPkt(unit)

	h.audioStart.Store(true)

	h.mu.Lock()
	now := time.Now()
	if trackCtx.prevTime.IsZero() {
		trackCtx.prevTime = now
	}
	if now.Sub(trackCtx.prevTime) >= 1*time.Second {
		trackCtx.prevTime = now

		buf := avformat.AVIOCloseDynBuf(h.outputFormatCtx.Pb())

		diff := pkt.PTS() - trackCtx.prevPTS
		trackCtx.prevPTS = pkt.PTS()
		fDuration := float64(diff) / float64(trackCtx.outputStream.TimeBase().Den())
		h.endpoint.AppendMedia(buf, h.index, fDuration)
		h.index++

		avioCtx := avformat.AVIOOpenDynBuf()
		h.outputFormatCtx.SetPb(avioCtx)
	}

	if h.outputFormatCtx.Pb() != nil {
		_ = h.outputFormatCtx.AvInterleavedWriteFrame(pkt)
	}

	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *OnTrackContext, unit units.Unit) error {
	if !h.audioStart.Load() {
		return nil
	}

	writer := trackCtx.writer
	pkt := writer.WriteAudioPkt(unit)

	h.mu.Lock()
	if h.outputFormatCtx.Pb() != nil {
		_ = h.outputFormatCtx.AvInterleavedWriteFrame(pkt)
	}
	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}
