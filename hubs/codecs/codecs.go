package codecs

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs/bitstreamfilter"
	"mediaserver-go/thirdparty/ffmpeg/avutil"

	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/utils/types"
)

var (
	errUnsupportedWebRTC = errors.New("unsupported webrtc codec")
)

type Codec interface {
	RTPCodecCapability

	Type() CodecType

	String() string
	Equals(codec Codec) bool
	Clone() Codec

	CodecType() types.CodecType
	MediaType() types.MediaType
	SetCodecContext(codecCtx *avcodec.CodecContext)

	WebRTCCodecCapability() (pion.RTPCodecCapability, error)

	GetBitStreamFilter() bitstreamfilter.BitStreamFilter
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

	SetVideoTranscodeInfo(info VideoTranscodeInfo)

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

type VideoTranscodeInfo struct {
	GOPSize       int
	FPS           int
	MaxBFrameSize int
}
