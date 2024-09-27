package codecs

import (
	"github.com/pion/rtp"
)

type RTPParser interface {
	Parse(rtpPacket *rtp.Packet) [][]byte
}
