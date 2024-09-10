package types

import (
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avutil"
	"strings"
)

// MediaType is the type of media
type MediaType string

const (
	UnknownMediaType MediaType = "unknown"
	MediaTypeVideo   MediaType = "video"
	MediaTypeAudio   MediaType = "audio"
)

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

func MediaTypeToFFMPEG(mediaType MediaType) avutil.MediaType {
	switch mediaType {
	case MediaTypeAudio:
		return avutil.AVMEDIA_TYPE_AUDIO
	case MediaTypeVideo:
		return avutil.AVMEDIA_TYPE_VIDEO
	default:
		return avutil.AVMEDIA_TYPE_UNKNOWN
	}
}

// CodecType is the type of codec
type CodecType string

const (
	CodecTypeUnknown CodecType = "unknown"
	CodecTypeH264    CodecType = "h264"
	CodecTypeVP8     CodecType = "vp8"
	CodecTypeAAC     CodecType = "aac"
	CodecTypeOpus    CodecType = "opus"
)

func CodecTypeFromFFMPEG(codecID avcodec.CodecID) CodecType {
	switch codecID {
	case avcodec.AV_CODEC_ID_H264:
		return CodecTypeH264
	case avcodec.AV_CODEC_ID_AAC:
		return CodecTypeAAC
	case avcodec.AV_CODEC_ID_OPUS:
		return CodecTypeOpus
	default:
		return CodecTypeUnknown
	}
}

func CodecIDFromType(codecType CodecType) avcodec.CodecID {
	switch codecType {
	case CodecTypeH264:
		return avcodec.AV_CODEC_ID_H264
	case CodecTypeAAC:
		return avcodec.AV_CODEC_ID_AAC
	case CodecTypeOpus:
		return avcodec.AV_CODEC_ID_OPUS
	default:
		return avcodec.AV_CODEC_ID_NONE
	}
}

func CodecTypeFromMimeType(mimeType string) CodecType {
	switch strings.ToLower(mimeType) {
	case "video/h264":
		return CodecTypeH264
	case "video/vp8":
		return CodecTypeVP8
	case "audio/aac":
		return CodecTypeAAC
	case "audio/opus":
		return CodecTypeOpus
	}
	return CodecTypeUnknown
}
