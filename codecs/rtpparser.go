package codecs

import (
	"github.com/pion/rtp"
	"mediaserver-go/utils/units"
)

type RTPParser interface {
	Parse(rtpPacket *rtp.Packet) ([][]byte, units.FrameInfo)
}
