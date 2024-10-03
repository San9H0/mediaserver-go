package av1

import "github.com/bluenviron/mediacommon/pkg/codecs/av1"

const (
	OBUTypeSequenceHeader = av1.OBUTypeSequenceHeader
)

type OBUHeader struct {
	av1.OBUHeader
}
