package hls

import (
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/writers"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avformat"
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
