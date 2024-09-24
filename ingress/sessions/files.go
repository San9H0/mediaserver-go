package sessions

import "C"
import (
	"context"
	"errors"
	"fmt"
	"io"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/codecs/bitstreamfilter"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"slices"
	"time"
)

var (
	errCreateFileSession = errors.New("failed to create file session")
)

type trackContext struct {
	hubSource       *hubs.HubSource
	codec           codecs.Codec
	bitStreamFilter bitstreamfilter.BitStreamFilter
}

type FileSession struct {
	live bool

	trackCtx [2]*trackContext

	inputFormatCtx *avformat.FormatContext

	stream *hubs.Stream
}

func NewFileSession(path string, mediaTypes []types.MediaType, live bool, hubStream *hubs.Stream) (FileSession, error) {
	var trackCtx [2]*trackContext

	inputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatOpenInput(&inputFormatCtx, path, nil, nil); ret < 0 {
		return FileSession{}, fmt.Errorf("avformat open input failed: %s, %w", avutil.AvErr2str(ret), errCreateFileSession)
	}
	if ret := inputFormatCtx.AvformatFindStreamInfo(nil); ret < 0 {
		return FileSession{}, fmt.Errorf("avformat find stream info failed: %s, %w", avutil.AvErr2str(ret), errCreateFileSession)
	}
	for i, stream := range inputFormatCtx.Streams() {
		codecType := types.CodecTypeFromFFMPEG(stream.CodecParameters().CodecID())
		mediaType := types.MediaTypeFromFFMPEG(stream.CodecParameters().CodecType())
		if !slices.Contains(mediaTypes, mediaType) {
			continue
		}

		codec, err := codecs.NewCodecFromAVStream(stream)
		if err != nil {
			return FileSession{}, err
		}

		source := hubs.NewHubSource(mediaType, codecType)
		hubStream.AddSource(source)
		source.SetCodec(codec)
		trackCtx[i] = &trackContext{
			hubSource:       source,
			codec:           codec,
			bitStreamFilter: codec.GetBitStreamFilter(),
		}

		switch mediaType {
		case types.MediaTypeAudio:
		case types.MediaTypeVideo:
			videoCodec := codec.(codecs.VideoCodec)
			videoCodec.SetVideoTranscodeInfo(codecs.VideoTranscodeInfo{
				GOPSize:       30,
				FPS:           30,
				MaxBFrameSize: 0,
			})
			source.SetTranscodeCodec(videoCodec)
		}
	}

	if trackCtx[0] == nil && trackCtx[1] == nil {
		return FileSession{}, fmt.Errorf("no audio or video stream found, %w", errCreateFileSession)
	}

	return FileSession{
		live:           live,
		trackCtx:       trackCtx,
		stream:         hubStream,
		inputFormatCtx: inputFormatCtx,
	}, nil
}

func (s *FileSession) Run(ctx context.Context) error {
	defer func() {
		s.inputFormatCtx.AvformatCloseInput()
		s.stream.Close()
	}()

	var startTime time.Time

	pkt := avcodec.AvPacketAlloc()
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

		index := pkt.StreamIndex()
		trackCtx := s.trackCtx[index]
		if trackCtx == nil {
			continue
		}

		if startTime.IsZero() {
			startTime = time.Now()
		}

		stream := s.inputFormatCtx.Streams()[pkt.StreamIndex()]
		itb := stream.TimeBase()
		ptsTimeSec := avutil.GetTimebaseUSec(itb, pkt.PTS())
		diffus := time.Now().Sub(startTime).Microseconds()
		delay := ptsTimeSec - (diffus)

		for _, au := range trackCtx.bitStreamFilter.Filter(pkt.Data()) {
			trackCtx.hubSource.Write(units.Unit{
				Payload:  au,
				PTS:      pkt.PTS(),
				DTS:      pkt.DTS(),
				Duration: pkt.Duration(),
				TimeBase: stream.TimeBase().Den(),
			})
		}

		pkt.AvPacketUnref()

		if types.MediaTypeFromFFMPEG(stream.CodecParameters().CodecType()) == types.MediaTypeVideo {
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Microsecond)
			}
		}
	}
}
