package codecs

import (
	"errors"
	"mediaserver-go/ffmpeg/goav/avutil"

	pion "github.com/pion/webrtc/v3"

	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils/types"
)

var (
	errUnsupportedWebRTC = errors.New("unsupported webrtc codec")
)

type Codec interface {
	String() string
	Equals(codec Codec) bool
	CodecType() types.CodecType
	MediaType() types.MediaType
	SetCodecContext(codecCtx *avcodec.CodecContext)

	WebRTCCodecCapability() (pion.RTPCodecCapability, error)
	RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error)
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
	ExtraData() []byte
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
