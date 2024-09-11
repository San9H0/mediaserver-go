package codecs

import "mediaserver-go/utils/types"

type Codec interface {
	CodecType() types.CodecType
	MediaType() types.MediaType
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
	FPS() float64
	PixelFormat() int
	ExtraData() []byte
	Profile() string
}
