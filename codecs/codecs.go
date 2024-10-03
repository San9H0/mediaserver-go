package codecs

import (
	"errors"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/thirdparty/ffmpeg/avutil"

	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/utils/types"
)

var (
	errUnsupportedCodec  = errors.New("unsupported codec")
	errUnsupportedWebRTC = errors.New("unsupported webrtc codec")
)

type BitStreamFilter interface {
	AddFilter(payload []byte) []byte
	Filter([]byte) [][]byte
}

type Codec interface {
	Base
	RTPCodecCapability

	String() string
	HLSMIME() string
	Equals(codec Codec) bool

	SetCodecContext(codecCtx *avcodec.CodecContext, transcodeInfo *VideoTranscodeInfo)

	WebRTCCodecCapability() (pion.RTPCodecCapability, error)

	GetBitStreamFilter(fromTranscoding bool) BitStreamFilter
	ExtraData() []byte
}

type AudioCodec interface {
	Codec

	AvCodecFifoAlloc() *avutil.AvAudioFifo
	SampleRate() int
	Channels() int
	SampleFormat() int
}

type VideoCodec interface {
	Codec

	Width() int
	Height() int
	ClockRate() uint32
	FPS() float64
	PixelFormat() int
}

func GetExtension(videoCodec VideoCodec, audioCodec AudioCodec) (string, error) {
	extension := ""
	if videoCodec != nil && audioCodec != nil {
		switch videoCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "webm"
		case types.CodecTypeH264, types.CodecTypeAV1:
			extension = "mp4"
		default:
			return "", fmt.Errorf("video codec:%v. %w", videoCodec.CodecType(), errUnsupportedCodec)
		}
	} else if videoCodec != nil {
		switch videoCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "mkv"
		case types.CodecTypeH264, types.CodecTypeAV1:
			extension = "m4v"
		default:
			return "", fmt.Errorf("video codec:%v. %w", videoCodec.CodecType(), errUnsupportedCodec)
		}
	} else if audioCodec != nil {
		switch audioCodec.CodecType() {
		case types.CodecTypeVP8:
			extension = "mka"
		case types.CodecTypeH264:
			extension = "m4a"
		default:
			return "", fmt.Errorf("audio codec:%v. %w", audioCodec.CodecType(), errUnsupportedCodec)
		}
	}
	return extension, nil
}

type VideoTranscodeInfo struct {
	GOPSize       int
	FPS           int
	MaxBFrameSize int
}
