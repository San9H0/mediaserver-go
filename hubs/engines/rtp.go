package engines

import (
	"github.com/pion/sdp/v3"
	"mediaserver-go/utils/types"
)

type RTPCodecParameters struct {
	MediaDescription sdp.MediaDescription
	CodecType        types.CodecType
	PayloadType      uint8
	ClockRate        uint32
}
