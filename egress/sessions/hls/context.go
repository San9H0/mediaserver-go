package hls

import (
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/writers"
	"time"
)

type OnTrackContext struct {
	track        *hubs.Track
	pkt          *avcodec.Packet
	outputStream *avformat.Stream
	writer       *writers.Writer
	count        int
	setup        bool
	prevTime     time.Time
	prevPTS      int64
}
