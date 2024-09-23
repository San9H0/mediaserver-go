package files

import (
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/hubs/writers"
)

type TrackContext struct {
	pkt          *avcodec.Packet
	outputStream *avformat.Stream
	writer       *writers.Writer
}
