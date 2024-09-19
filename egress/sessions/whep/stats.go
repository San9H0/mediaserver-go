package whep

import "sync/atomic"

type Stats struct {
	lastTS     atomic.Uint32
	sendCount  atomic.Uint32
	sendLength atomic.Uint32
}
