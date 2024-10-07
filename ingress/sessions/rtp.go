package sessions

import (
	"context"
	"fmt"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/factory"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions/rtpinbounder"
	"mediaserver-go/utils/types"
	"net"
)

type RTPSession struct {
	conn      *net.UDPConn
	stream    *hubs.Stream
	pt        uint8
	hubSource *hubs.HubSource
	timebase  int

	codecType codecs.Base
}

func NewRTPSession(ip string, port int, pt uint8, mimeType string, stream *hubs.Stream) (RTPSession, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return RTPSession{}, err
	}

	base, err := factory.NewBase(mimeType)
	if err != nil {
		return RTPSession{}, err
	}

	hubSource := hubs.NewHubSource(base, "")
	timebase := 0

	switch types.CodecTypeFromMimeType(mimeType) {
	case types.CodecTypeH264:
		timebase = 90000
	case types.CodecTypeVP8:
		timebase = 90000
	case types.CodecTypeOpus:
		timebase = 48000
	default:
		return RTPSession{}, fmt.Errorf("unsupported codec type: %v", mimeType)
	}

	stream.AddSource(hubSource)

	return RTPSession{
		pt:        pt,
		conn:      conn,
		stream:    stream,
		hubSource: hubSource,
		timebase:  timebase,

		codecType: base,
	}, nil
}

func (r *RTPSession) Run(ctx context.Context) error {
	parser, err := r.codecType.RTPParser(func(codec codecs.Codec) {
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
