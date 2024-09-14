package sessions

import (
	"context"
	"fmt"
	"github.com/pion/sdp/v3"
	"golang.org/x/sync/errgroup"
	"mediaserver-go/egress/sessions/packetizers"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/types"
	"net"
)

type CodecInfo struct {
	PayloadType uint8
	ClockRate   uint32
}

type RTPSession struct {
	conn         *net.UDPConn
	sourceTracks []*hubs.Track

	targetAddr string
	targetPort int
	sd         sdp.SessionDescription
}

func NewRTPSession(targetAddr string, targetPort int, sourceTracks []*hubs.Track) (*RTPSession, error) {
	target, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetAddr, targetPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, target)
	if err != nil {
		return nil, err
	}

	sd := makeSessionDescription(targetAddr)
	for _, sourceTrack := range sourceTracks {
		codec, err := sourceTrack.Codec()
		if err != nil {
			return nil, err
		}
		capa, err := codec.RTPCodecCapability(targetPort)
		if err != nil {
			return nil, err
		}
		sd.MediaDescriptions = append(sd.MediaDescriptions, &capa.MediaDescription)
	}

	return &RTPSession{
		targetAddr:   targetAddr,
		targetPort:   targetPort,
		conn:         conn,
		sourceTracks: sourceTracks,
		sd:           sd,
	}, nil
}

func (r *RTPSession) SDP() string {
	b, err := r.sd.Marshal()
	if err != nil {
		return ""
	}
	fmt.Println(string(b))
	return string(b)
}

func (r *RTPSession) Run(ctx context.Context) error {
	defer r.conn.Close()
	g, ctx := errgroup.WithContext(ctx)
	for _, track := range r.sourceTracks {
		g.Go(func() error {
			return r.readTrack(ctx, track)
		})
	}

	return g.Wait()
}

// 현재는 h264만 지원.
func (r *RTPSession) readTrack(ctx context.Context, track *hubs.Track) error {
	consumerCh := track.AddConsumer()
	defer func() {
		track.RemoveConsumer(consumerCh)
	}()
	codec, err := track.Codec()
	if err != nil {
		return err
	}
	rtpCapability, err := codec.RTPCodecCapability(r.targetPort)
	if err != nil {
		return err
	}
	packetizer, err := packetizers.NewPacketizer(rtpCapability, codec)
	if err != nil {
		return err
	}
	buf := make([]byte, types.ReadBufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		case unit, ok := <-consumerCh:
			if !ok {
				return nil
			}
			for _, rtpPacket := range packetizer.Packetize(unit) {
				n, err := rtpPacket.MarshalTo(buf)
				if err != nil {
					continue
				}
				if _, err := r.conn.Write(buf[:n]); err != nil {
					fmt.Println("error writing to conn:", err)
				}
			}
		}
	}
}

func makeSessionDescription(targetAddr string) sdp.SessionDescription {
	return sdp.SessionDescription{
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      0,
			SessionVersion: 0,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: targetAddr,
		},
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address: &sdp.Address{
				Address: targetAddr,
			},
		},
		TimeDescriptions: []sdp.TimeDescription{
			{
				Timing: sdp.Timing{
					StartTime: 0,
					StopTime:  0,
				},
			},
		},
	}
}
