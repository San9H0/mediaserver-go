package codecs

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/utils/types"
)

var (
	errUnsupportedWebRTC = errors.New("unsupported webrtc codec")
)

type Codec interface {
	CodecType() types.CodecType
	MediaType() types.MediaType
	WebRTCCodecCapability() (pion.RTPCodecCapability, error)
	SetCodecContext(codecCtx *avcodec.CodecContext)
}

type AudioCodec interface {
	Codec

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
