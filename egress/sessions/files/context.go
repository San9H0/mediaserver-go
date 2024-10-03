package files

import (
	"mediaserver-go/codecs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
)

type TrackContext struct {
	codec        codecs.Codec
	outputStream *avformat.Stream
	writer       *writers.Writer
}
