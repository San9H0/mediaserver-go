package files

import (
	"mediaserver-go/hubs/writers"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
)

type TrackContext struct {
	pkt          *avcodec.Packet
	outputStream *avformat.Stream
	writer       *writers.Writer
}
