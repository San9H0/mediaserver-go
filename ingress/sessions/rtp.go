package sessions

import (
	"context"
	"fmt"
	"github.com/pion/rtp"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/parsers"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"net"
)

type RTPSession struct {
	conn     *net.UDPConn
	stream   *hubs.Stream
	pt       uint8
	track    *hubs.Track
	parser   parsers.Parser
	timebase int
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

	parser, err := parsers.NewParser(codecType)
	if err != nil {
		return RTPSession{}, err
	}

	fmt.Println("[TESTDEBG] codecParser:", parser)
	var track *hubs.Track
	timebase := 0
	switch codecType {
	case types.CodecTypeH264:
		track = hubs.NewTrack(types.MediaTypeVideo, types.CodecTypeH264)
		stream.AddTrack(track)
		timebase = 90000
	case types.CodecTypeVP8:
		track = hubs.NewTrack(types.MediaTypeVideo, types.CodecTypeVP8)
		stream.AddTrack(track)
		timebase = 90000
	case types.CodecTypeOpus:
		track = hubs.NewTrack(types.MediaTypeAudio, types.CodecTypeOpus)
		stream.AddTrack(track)
		timebase = 48000
	default:
		return RTPSession{}, fmt.Errorf("unsupported codec type: %v", codecType)
	}

	return RTPSession{
		pt:       pt,
		conn:     conn,
		stream:   stream,
		parser:   parser,
		track:    track,
		timebase: timebase,
	}, nil
}

func (r *RTPSession) Run(ctx context.Context) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0
	var codec codecs.Codec
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		buf := make([]byte, types.ReadBufferSize)

		n, _, err := r.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("error reading from udp:", err)
			return err
		}

		rtpPacket := &rtp.Packet{}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			fmt.Println("error unmarshalling rtp packet:", err)
			continue
		}

		if startTS == 0 {
			startTS = rtpPacket.Timestamp
		}
		pts := rtpPacket.Timestamp - startTS

		if rtpPacket.Timestamp != prevTS {
			if prevTS == 0 {
				duration = 0
			} else {
				duration = int(rtpPacket.Timestamp - prevTS)
			}
		}

		aus := r.parser.Parse(rtpPacket)
		newCodec := r.parser.GetCodec()
		if newCodec == nil {
			continue
		} else if newCodec != codec {
			r.track.SetCodec(newCodec)
			codec = newCodec
		}

		for _, au := range aus {
			r.track.Write(units.Unit{
				Payload:  au,
				PTS:      int64(pts),
				DTS:      int64(pts),
				Duration: int64(duration),
				TimeBase: r.timebase,
			})

			//packetizers.CommonPacketizer{}
		}
	}
}
