package whep

import (
	"sync/atomic"
)

type Stats struct {
	lastNTP    atomic.Uint64
	lastTS     atomic.Uint32
	sendCount  atomic.Uint32
	sendLength atomic.Uint32

	nackCount atomic.Uint32
}

func NewStats() *Stats {
	return &Stats{}
}
