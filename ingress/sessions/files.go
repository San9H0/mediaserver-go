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
	"mediaserver-go/parser/format"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"slices"
	"time"
)

const (
	bufferSize = 32768
)

type FileSession struct {
	live bool

	audioIndex, videoIndex int
	audioTrack, videoTrack *hubs.Track

	inputFormatCtx *avformat.FormatContext

	stream *hubs.Stream
}

func NewFileSession(path string, mediaTypes []types.MediaType, live bool, stream *hubs.Stream) (FileSession, error) {
	audioIndex, videoIndex := -1, -1
	var audioTrack, videoTrack *hubs.Track
	inputFormatCtx := avformat.NewAvFormatContextNull()
	if ret := avformat.AvformatOpenInput(&inputFormatCtx, path, nil, nil); ret < 0 {
		return FileSession{}, errors.New("avformat open input failed")
	}
	if ret := inputFormatCtx.AvformatFindStreamInfo(nil); ret < 0 {
		return FileSession{}, errors.New("avformat find stream info failed")
	}
	fmt.Println("inputFormatCtx.NbStreams():", inputFormatCtx.NbStreams())
	tracks := make(map[int]types.Track)
	for i := 0; i < int(inputFormatCtx.NbStreams()); i++ {
		inputStream := inputFormatCtx.Streams()[i]
		codecType := types.CodecTypeFromFFMPEG(inputStream.CodecParameters().CodecID())
		mediaType := types.MediaTypeFromFFMPEG(inputStream.CodecParameters().CodecType())
		fmt.Println("codecType:", codecType, ", mediaType:", mediaType)
		if !slices.Contains(mediaTypes, mediaType) {
			continue
		}

		if mediaType == types.MediaTypeVideo {
			videoIndex = i
			videoTrack = hubs.NewTrack(types.MediaTypeVideo, codecType)
			stream.AddTrack(videoTrack)
		} else if mediaType == types.MediaTypeAudio {
			audioIndex = i
			audioTrack = hubs.NewTrack(types.MediaTypeAudio, codecType)
			stream.AddTrack(audioTrack)
		}
		tracks[i] = types.NewTrack(mediaType, codecType)
	}
	if audioIndex == -1 && videoIndex == -1 {
		return FileSession{}, fmt.Errorf("no audio or video stream found audioIndex:%d, videoIndex:%d", audioIndex, videoIndex)
	}

	fmt.Println("[TESTDEBUG] videoTrack != nil:", videoTrack != nil)
	if videoTrack != nil {
		inputStream := inputFormatCtx.Streams()[videoIndex]

		inputCodecpar := inputStream.CodecParameters()
		sps, pps := format.SPSPPSFromAVCCExtraData(inputCodecpar.ExtraData())
		if len(sps) == 0 || len(pps) == 0 {
			return FileSession{}, errors.New("sps pps not found")
		}

		h264Codecs, err := codecs.NewH264(sps, pps)
		if err != nil {
			return FileSession{}, err
		}

		videoTrack.SetCodec(h264Codecs)
	}

	if audioTrack != nil {
		inputStream := inputFormatCtx.Streams()[audioIndex]
		inputCodecpar := inputStream.CodecParameters()

		fmt.Println("[TESTDEBUG] aac sampleRate:", inputCodecpar.SampleRate(), ", channels:", inputCodecpar.Channels(), ", sampleFmt:", inputCodecpar.Format())
		audioTrack.SetCodec(codecs.NewAAC(codecs.AACParameters{
			SampleRate: inputCodecpar.SampleRate(),
			Channels:   inputCodecpar.Channels(),
			SampleFmt:  inputCodecpar.Format(),
		}))
	}

	fmt.Println("videoIndex:", videoIndex, ", audioIndex:", audioIndex)

	return FileSession{
		live:           live,
		audioIndex:     audioIndex,
		videoIndex:     videoIndex,
		audioTrack:     audioTrack,
		videoTrack:     videoTrack,
		stream:         stream,
		inputFormatCtx: inputFormatCtx,
	}, nil
}

func (s *FileSession) Run(ctx context.Context) error {
	defer func() {
		fmt.Println("[TESTDEBUG] ingress file session closed")
		s.inputFormatCtx.AvformatCloseInput()
		fmt.Println("file session done")
		s.Finish()
		fmt.Println("file write done")
		s.stream.Close()
	}()

	fmt.Println("[TESTDEBUG] fileSession run")

	var startTime time.Time

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		pkt := avcodec.AvPacketAlloc()
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
			for _, au := range format.GetAUFromAVC(pkt.Data()) {
				s.videoTrack.Write(units.Unit{
					Payload:  au,
					PTS:      pkt.PTS(),
					DTS:      pkt.DTS(),
					Duration: pkt.Duration(),
					TimeBase: istream.TimeBase().Den(),
				})
			}

			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Microsecond)
			}
		} else if pkt.StreamIndex() == s.audioIndex {
			itb := s.inputFormatCtx.Streams()[pkt.StreamIndex()].TimeBase()
			ptsTimeSec := avutil.GetTimebaseUSec(itb, pkt.PTS())
			diffus := time.Now().Sub(startTime).Microseconds()
			delay := ptsTimeSec - (diffus)

			istream := s.inputFormatCtx.Streams()[s.audioIndex]
			s.audioTrack.Write(units.Unit{
				Payload:  pkt.Data(),
				PTS:      pkt.PTS(),
				DTS:      pkt.DTS(),
				Duration: pkt.Duration(),
				TimeBase: istream.TimeBase().Den(),
			})
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Microsecond)
			}
		}

	}
}

func (s *FileSession) Finish() {
	return
}
