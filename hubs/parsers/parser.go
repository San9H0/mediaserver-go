package parsers

import "github.com/pion/rtp"

type Parser interface {
	Parse(payload *rtp.Packet) [][]byte
}
