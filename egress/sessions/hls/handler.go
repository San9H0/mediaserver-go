package hls

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mediaserver-go/ffmpeg/goav/avutil"
	"os"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"

	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Endpoint interface {
	SetPayload(payload []byte, name string)
	AppendMedia(payload []byte, name string, duration float64)
}

type Handler struct {
	mu sync.RWMutex

	ioBuffer io.ReadWriteSeeker
	endpoint Endpoint

	extension       string
	negotidated     []*hubs.Track
	outputFormatCtx *avformat.FormatContext
	index           int
}

func NewHandler(buffer io.ReadWriteSeeker, endpoint Endpoint) *Handler {
	return &Handler{
		ioBuffer: buffer,
		endpoint: endpoint,
	}
}

func (h *Handler) CodecString(mediaType types.MediaType) string {
	for i, stream := range h.outputFormatCtx.Streams() {
		if types.MediaTypeFromFFMPEG(stream.CodecParameters().CodecType()) != mediaType {
			continue
		}
		codecType := types.CodecTypeFromFFMPEG(stream.CodecParameters().CodecID())
		switch codecType {
		case types.CodecTypeH264:
			str := avutil.AvFourcc2str(stream.CodecParameters().CodecTag())
			codec, err := h.negotidated[i].VideoCodec()
			if err != nil {
				continue
			}
			profile := stream.CodecParameters().Profile()
			constraint_flags := codec.ExtraData()[2]
			level := stream.CodecParameters().Level()
			return fmt.Sprintf("%s.%02X%02X%02x",
				str,
				profile, constraint_flags, level)
		case types.CodecTypeAAC:
			str := avutil.AvFourcc2str(stream.CodecParameters().CodecTag())
			audioObjectType := 2
			return fmt.Sprintf("%s.40.%d", str, audioObjectType)
		case types.CodecTypeOpus:
			str := avutil.AvFourcc2str(stream.CodecParameters().CodecTag())
			return fmt.Sprintf("%s", str)
		}
	}
	return ""
}

func (h *Handler) NegotiatedTracks() []*hubs.Track {
	ret := make([]*hubs.Track, 0, len(h.negotidated))
	return append(ret, h.negotidated...)
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

	if videoCodec.CodecType() != types.CodecTypeH264 {
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

	for _, sourceTrack := range negotiated {
		sourceCodec, err := sourceTrack.Codec()
		if err != nil {
			return fmt.Errorf("codec not found: %w", err)
		}
		outputStream := h.outputFormatCtx.AvformatNewStream(nil)
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
		fmt.Println("[TESTDEBUG] CodecID:", outputStream.CodecParameters().CodecID())
		fmt.Println("[TESTDEBUG] CodecTag:", outputStream.CodecParameters().CodecTag(), ", str:", avutil.AvFourcc2str(outputStream.CodecParameters().CodecTag()))
		fmt.Println("[TESTDEBUG] Profile:", outputStream.CodecParameters().Profile())
		fmt.Println("[TESTDEBUG] Level:", outputStream.CodecParameters().Level())
	}

	h.outputFormatCtx.SetPb(avformat.AVIOOpenDynBuf())

	dict := avutil.DictionaryNull()
	avutil.AvDictSet(&dict, "movflags", "frag_keyframe+empty_moov+default_base_moof", 0)

	if outputFormatCtx.Flags()&0x0001 == 0 {

	}

	if ret := outputFormatCtx.AvformatWriteHeader(&dict); ret < 0 {
		return errors.New("avformat write header failed")
	}

	buf := avformat.AVIOCloseDynBuf(h.outputFormatCtx.Pb())
	h.endpoint.SetPayload(buf, "init.mp4")

	h.outputFormatCtx.SetPb(avformat.AVIOOpenDynBuf())

	h.extension = extension
	h.negotidated = negotiated
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	log.Logger.Info("hls Handler finish start")

	// TODO 남아있는 것 모두 정리.

	//avformat.AvIoContextFree(h.outputFormatCtx.Pb())
	//h.outputFormatCtx.AvformatFreeContext()

	//ioBuffer := h.ioBuffer
	//h.ioBuffer = nil
	//size, err := ioBuffer.Seek(0, io.SeekEnd)
	//if err != nil {
	//	return fmt.Errorf("error seeking to end of buffer: %w", err)
	//}
	//_, _ = ioBuffer.Seek(0, io.SeekStart)
	//
	//filepath := fmt.Sprintf("%s.%s", h.path, h.extension)
	//file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	//if err != nil {
	//	return fmt.Errorf("error opening file: %w", err)
	//}
	//defer file.Close()
	//if _, err := io.Copy(file, ioBuffer); err != nil {
	//	return fmt.Errorf("error copying to file: %w", err)
	//}
	//log.Logger.Info("file session is finished",
	//	zap.String("filepath", filepath),
	//	zap.Int("size", int(size)))
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track *hubs.Track) (*OnTrackContext, error) {
	index := slices.Index(h.negotidated, track)
	stream := h.outputFormatCtx.Streams()[index]
	return &OnTrackContext{
		track:        track,
		pkt:          avcodec.AvPacketAlloc(),
		outputStream: stream,
		writer:       writers.NewWriter(index, stream.TimeBase().Den(), track.CodecType()),
		prevTime:     time.Now(),
	}, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *OnTrackContext, unit units.Unit) error {
	writer := trackCtx.writer
	pkt := trackCtx.pkt
	setPkt := writer.WriteVideoPkt(unit, pkt)
	if setPkt == nil {
		return nil
	}

	h.mu.Lock()
	if setPkt.Flag() == 1 {
		if trackCtx.setup {
			now := time.Now()
			if now.Sub(trackCtx.prevTime) > 2*time.Second {
				trackCtx.prevTime = now
				buf := avformat.AVIOCloseDynBuf(h.outputFormatCtx.Pb())
				diff := setPkt.PTS() - trackCtx.prevPTS
				trackCtx.prevPTS = setPkt.PTS()
				fDuration := float64(diff) / float64(trackCtx.outputStream.TimeBase().Den())
				h.endpoint.AppendMedia(buf, fmt.Sprintf("output_%d.m4s", h.index), fDuration)

				filepath := fmt.Sprintf("output_%d.m4s", h.index)
				if err := writeFile(filepath, bytes.NewReader(buf)); err != nil {
					return fmt.Errorf("error writing file: %w", err)
				}
				h.index++

				avioCtx := avformat.AVIOOpenDynBuf()
				h.outputFormatCtx.SetPb(avioCtx)
			}
		} else {
			trackCtx.setup = true
		}
	}

	if h.outputFormatCtx.Pb() != nil {
		_ = h.outputFormatCtx.AvInterleavedWriteFrame(setPkt)
	}

	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *OnTrackContext, unit units.Unit) error {
	writer := trackCtx.writer
	pkt := trackCtx.pkt
	setPkt := writer.WriteAudioPkt(unit, pkt)
	if setPkt == nil {
		return nil
	}

	h.mu.Lock()
	if h.outputFormatCtx.Pb() != nil {
		_ = h.outputFormatCtx.AvInterleavedWriteFrame(setPkt)
	}
	h.mu.Unlock()
	pkt.AvPacketUnref()
	return nil
}

func writeFile(filepath string, reader io.Reader) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()
	size, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error copying to file: %w", err)
	}
	log.Logger.Info("hls Handler write file end",
		zap.String("filepath", filepath),
		zap.Int("size", int(size)))
	return nil
}
