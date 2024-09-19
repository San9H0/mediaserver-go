package rtp

import "mediaserver-go/egress/sessions/packetizers"

type TrackContext struct {
	packetizer packetizers.Packetizer
	buf        []byte
}
