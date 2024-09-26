package containers

import (
	"errors"
	"sync"

	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
)

type MP4 struct {
	mu sync.RWMutex

	video codecs.VideoCodec
	audio codecs.AudioCodec
}

func NewMP4(video codecs.VideoCodec, audio codecs.AudioCodec) *MP4 {
	return &MP4{
		video: video,
		audio: audio,
	}
}

func (m *MP4) Codecs() []codecs.Codec {
	var ret []codecs.Codec
	if m.video != nil {
		ret = append(ret, m.video)
	}
	if m.audio != nil {
		ret = append(ret, m.audio)
	}
	return ret
}

func (m *MP4) Extension() string {
	return "mp4"
}

func (m *MP4) SetWriteHeader(ctx *avformat.FormatContext) error {
	if ret := ctx.AvformatWriteHeaderWithFMP4("movflags", "frag_keyframe+empty_moov+default_base_moof"); ret < 0 {
		return errors.New("avformat write header failed for fmp4")
	}
	return nil
}
