package containers

import (
	"errors"

	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
)

type WebM struct {
	video codecs.VideoCodec
	audio codecs.AudioCodec
}

func NewWebM(video codecs.VideoCodec, audio codecs.AudioCodec) *WebM {
	return &WebM{
		video: video,
		audio: audio,
	}
}

func (w *WebM) Extension() string {
	return "webm"
}

func (w *WebM) Codecs() []codecs.Codec {
	var ret []codecs.Codec
	if w.video != nil {
		ret = append(ret, w.video)
	}
	if w.audio != nil {
		ret = append(ret, w.audio)
	}
	return ret
}

func (w *WebM) SetWriteHeader(ctx *avformat.FormatContext) error {
	if ret := ctx.AvformatWriteHeader(nil); ret < 0 {
		return errors.New("avformat write header failed")
	}
	return nil
}
