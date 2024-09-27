package rtp

import (
	"context"
	"fmt"
	"github.com/pion/sdp/v3"
	"mediaserver-go/egress/sessions/packetizers"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"net"
)

type CodecInfo struct {
	PayloadType uint8
	ClockRate   uint32
}

type Handler struct {
	conn         *net.UDPConn
	sourceTracks []*hubs.HubSource

	targetAddr  string
	targetPort  int
	sd          sdp.SessionDescription
	negotidated []hubs.Track
}

func NewHandler(targetAddr string, targetPort int) *Handler {
	return &Handler{
		targetAddr: targetAddr,
		targetPort: targetPort,
	}
}

func (h *Handler) NegotiatedTracks() []hubs.Track {
	ret := make([]hubs.Track, 0, len(h.negotidated))
	return append(ret, h.negotidated...)
}

func (h *Handler) SDP() string {
	b, err := h.sd.Marshal()
	if err != nil {
		return ""
	}
	fmt.Println(string(b))
	return string(b)
}

func (h *Handler) Init(ctx context.Context, sources []*hubs.HubSource) error {
	target, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", h.targetAddr, h.targetPort))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, target)
	if err != nil {
		return err
	}
	var negotidated []hubs.Track
	sd := h.makeSessionDescription()
	for _, source := range sources {
		codec, err := source.Codec()
		if err != nil {
			return err
		}
		capa, err := codec.RTPCodecCapability(h.targetPort)
		if err != nil {
			return err
		}
		sd.MediaDescriptions = append(sd.MediaDescriptions, &capa.MediaDescription)
		track := source.GetTrack(codec)
		negotidated = append(negotidated, track)
	}

	h.conn = conn
	h.sd = sd
	h.negotidated = negotidated
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track hubs.Track) (*TrackContext, error) {
	codec := track.GetCodec()
	rtpCapability, err := codec.RTPCodecCapability(h.targetPort)
	if err != nil {
		return nil, err
	}
	packetizer, err := packetizers.NewPacketizer(rtpCapability, codec)
	if err != nil {
		return nil, err
	}

	return &TrackContext{
		packetizer: packetizer,
		buf:        make([]byte, types.ReadBufferSize),
	}, nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	h.conn.Close()
	return nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	packetizer := trackCtx.packetizer
	buf := trackCtx.buf
	for _, rtpPacket := range packetizer.Packetize(unit) {
		n, err := rtpPacket.MarshalTo(buf)
		if err != nil {
			continue
		}
		if _, err := h.conn.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	packetizer := trackCtx.packetizer
	buf := trackCtx.buf
	for _, rtpPacket := range packetizer.Packetize(unit) {
		n, err := rtpPacket.MarshalTo(buf)
		if err != nil {
			continue
		}
		if _, err := h.conn.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) makeSessionDescription() sdp.SessionDescription {
	return sdp.SessionDescription{
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      0,
			SessionVersion: 0,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: h.targetAddr,
		},
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address: &sdp.Address{
				Address: h.targetAddr,
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
