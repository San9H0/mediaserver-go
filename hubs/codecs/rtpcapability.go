package codecs

import (
	"mediaserver-go/hubs/engines"
)

type RTPCodecCapability interface {
	RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error)
}
