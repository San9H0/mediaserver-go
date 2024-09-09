package types

import (
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/goav/avcodec"
	"mediaserver-go/goav/avutil"
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

// CodecType is the type of codec
type CodecType string

const (
	UnknownCodecType CodecType = "unknown"
	CodecTypeH264    CodecType = "h264"
	CodecTypeOpus    CodecType = "opus"
)

func CodecTypeFromFFMPEG(codecID avcodec.CodecID) CodecType {
	switch codecID {
	case avcodec.AV_CODEC_ID_H264:
		return CodecTypeH264
	case avcodec.AV_CODEC_ID_OPUS:
		return CodecTypeOpus
	default:
		return UnknownCodecType
	}
}
