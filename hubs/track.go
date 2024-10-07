package hubs

import (
	"mediaserver-go/codecs"
	"mediaserver-go/hubs/tracks"
	"mediaserver-go/utils/units"
)

type Track interface {
	Close()
	RID() string
	GetStats() *tracks.Stats
	InputCh() chan units.Unit
	GetCodec() codecs.Codec
	AddConsumer() chan units.Unit
	RemoveConsumer(consumerCh chan units.Unit)
}
