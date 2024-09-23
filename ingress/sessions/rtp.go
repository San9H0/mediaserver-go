package sessions

import (
	"context"
	"fmt"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/ingress/sessions/rtpinbounder"
	parsers2 "mediaserver-go/parsers"
	"mediaserver-go/utils/types"
	"net"
)

type RTPSession struct {
	conn      *net.UDPConn
	stream    *hubs.Stream
	pt        uint8
	hubSource *hubs.HubSource
	timebase  int

	codecConfig codecs.Config
}

func NewRTPSession(ip string, port int, pt uint8, codecType types.CodecType, stream *hubs.Stream) (RTPSession, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return RTPSession{}, err
	}

	codecConfig, err := codecs.NewCodecConfig(codecType)
	if err != nil {
		return RTPSession{}, err
	}
	var hubSource *hubs.HubSource
	timebase := 0
	switch codecType {
	case types.CodecTypeH264:
		hubSource = hubs.NewHubSource(types.MediaTypeVideo, types.CodecTypeH264)
		stream.AddSource(hubSource)
		timebase = 90000
	case types.CodecTypeVP8:
		hubSource = hubs.NewHubSource(types.MediaTypeVideo, types.CodecTypeVP8)
		stream.AddSource(hubSource)
		timebase = 90000
	case types.CodecTypeOpus:
		hubSource = hubs.NewHubSource(types.MediaTypeAudio, types.CodecTypeOpus)
		stream.AddSource(hubSource)
		timebase = 48000
	default:
		return RTPSession{}, fmt.Errorf("unsupported codec type: %v", codecType)
	}

	return RTPSession{
		pt:        pt,
		conn:      conn,
		stream:    stream,
		hubSource: hubSource,
		timebase:  timebase,

		codecConfig: codecConfig,
	}, nil
}

func (r *RTPSession) Run(ctx context.Context) error {
	parser, err := parsers2.NewRTPParser(r.codecConfig, func(codec codecs.Codec) {
		r.hubSource.SetCodec(codec)
	})
	if err != nil {
		return err
	}
	inbounder := rtpinbounder.NewInbounder(parser, r.timebase, func(bytes []byte) (int, error) {
		n, _, err := r.conn.ReadFromUDP(bytes)
		return n, err
	})
	stats := rtpinbounder.Stats{}
	return inbounder.Run(ctx, r.hubSource, &stats)
}
