package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/parsers"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"net"
	"sync"
)

type RTPSession struct {
	conn   *net.UDPConn
	stream *hubs.Stream
	ssrc   uint32
	pt     int
	tracks []*Track
}

type Track struct {
	target     *hubs.Track
	rtpTrackCh chan *rtp.Packet
	ssrc       uint32
	pt         uint8
}

func NewRTPSession(ip string, port int, ssrc uint32, pt uint8, stream *hubs.Stream) (RTPSession, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return RTPSession{}, err
	}

	var tracks []*Track
	target := hubs.NewTrack(types.MediaTypeVideo, types.CodecTypeH264)
	stream.AddTrack(target)

	tracks = append(tracks, &Track{
		target:     target,
		rtpTrackCh: make(chan *rtp.Packet, 100),
		ssrc:       ssrc,
		pt:         pt,
	})

	return RTPSession{
		conn:   conn,
		stream: stream,
		tracks: tracks,
	}, nil
}

func (r *RTPSession) Run(ctx context.Context) error {
	ssrcMap := map[uint32]*Track{}
	ptMap := map[uint8]*Track{}
	for _, track := range r.tracks {
		ssrcMap[track.ssrc] = track
		ptMap[track.pt] = track
		go r.readRTP(ctx, track)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		buf := make([]byte, 3000)

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

		track, ok := ptMap[rtpPacket.PayloadType]
		if !ok {
			continue
		}

		utils.SendOrDrop(track.rtpTrackCh, rtpPacket)
	}
}

func (r *RTPSession) readRTP(ctx context.Context, track *Track) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0
	parser := parsers.NewH264Parser()
	var once sync.Once
	for {
		select {
		case <-ctx.Done():
			return nil
		case rtpPacket := <-track.rtpTrackCh:
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

			aus := parser.Parse(rtpPacket.Payload)
			codec := parser.GetCodec()
			if codec == nil {
				continue
			}
			once.Do(func() {
				track.target.SetVideoCodec(codec)
			})

			for _, au := range aus {
				naluType := h264.NALUType(au[0] & 0x1F)
				flags := 0
				if naluType == h264.NALUTypeIDR {
					flags = 1
				}
				track.target.Write(units.Unit{
					Payload:  au,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: 90000,
					Flags:    flags,
				})
			}
		}
	}
}
