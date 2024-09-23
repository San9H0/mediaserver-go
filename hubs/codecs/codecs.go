package codecs

import (
	"errors"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/parsers/format"

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

	BitStreamFilter([]byte) [][]byte
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

func NewCodecFromAVStream(stream *avformat.Stream) (Codec, error) {
	param := stream.CodecParameters()
	switch types.CodecTypeFromFFMPEG(stream.CodecParameters().CodecID()) {
	case types.CodecTypeH264:
		sps, pps := format.SPSPPSFromAVCCExtraData(param.ExtraData())
		if len(sps) == 0 || len(pps) == 0 {
			return nil, errors.New("sps pps not found")
		}
		h264Codecs, err := NewH264(sps, pps)
		if err != nil {
			return nil, err
		}
		return h264Codecs, nil
	case types.CodecTypeVP8:
	case types.CodecTypeOpus:
	case types.CodecTypeAAC:
		return NewAAC(AACParameters{
			SampleRate: param.SampleRate(),
			Channels:   param.Channels(),
			SampleFmt:  param.Format(),
		}), nil
	}
	return nil, nil
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
