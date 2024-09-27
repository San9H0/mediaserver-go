package hubs

import (
	"mediaserver-go/codecs"
	"mediaserver-go/utils/units"
)

type Track interface {
	Close()
	InputCh() chan units.Unit
	GetCodec() codecs.Codec
	AddConsumer() chan units.Unit
	RemoveConsumer(consumerCh chan units.Unit)
}
