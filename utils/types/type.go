package types

import (
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"strings"
)

// MediaType is the type of media
type MediaType string

const (
	UnknownMediaType MediaType = "unknown"
	MediaTypeVideo   MediaType = "video"
	MediaTypeAudio   MediaType = "audio"
)

func (m MediaType) String() string {
	return string(m)
}

func MediaTypeFromFFMPEG(mediaType avutil.MediaType) MediaType {
	switch mediaType {
	case avutil.AVMEDIA_TYPE_AUDIO:
		return MediaTypeAudio
	case avutil.AVMEDIA_TYPE_VIDEO:
		return MediaTypeVideo
	default:
		return UnknownMediaType
	}
}

func MediaTypeFromPion(mediaType pion.RTPCodecType) MediaType {
	switch mediaType {
	case pion.RTPCodecTypeAudio:
		return MediaTypeAudio
	case pion.RTPCodecTypeVideo:
		return MediaTypeVideo
	default:
		return UnknownMediaType
	}
}

// CodecType is the type of codec
type CodecType string

const (
	CodecTypeUnknown CodecType = "unknown"
	CodecTypeH264    CodecType = "h264"
	CodecTypeVP8     CodecType = "vp8"
	CodecTypeAV1     CodecType = "av1"
	CodecTypeAAC     CodecType = "aac"
	CodecTypeOpus    CodecType = "opus"
)

func CodecTypeFromFFMPEG(codecID avcodec.CodecID) CodecType {
	switch codecID {
	case avcodec.AV_CODEC_ID_H264:
		return CodecTypeH264
	case avcodec.AV_CODEC_ID_VP8:
		return CodecTypeVP8
	case avcodec.AV_CODEC_ID_AV1:
		return CodecTypeAV1
	case avcodec.AV_CODEC_ID_AAC:
		return CodecTypeAAC
	case avcodec.AV_CODEC_ID_OPUS:
		return CodecTypeOpus
	default:
		return CodecTypeUnknown
	}
}

func CodecTypeFromMimeType(mimeType string) CodecType {
	switch strings.ToLower(mimeType) {
	case "video/h264":
		return CodecTypeH264
	case "video/vp8":
		return CodecTypeVP8
	case "video/av1":
		return CodecTypeAV1
	case "audio/aac":
		return CodecTypeAAC
	case "audio/opus":
		return CodecTypeOpus
	}
	return CodecTypeUnknown
}
