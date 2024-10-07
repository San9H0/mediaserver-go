package tracks

import (
	"context"
	"mediaserver-go/utils/units"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	mu sync.RWMutex

	totalBytes atomic.Uint32
	bitrate    atomic.Uint32
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) Run(ctx context.Context) {
	per := time.NewTicker(time.Second)
	prevBytes := uint32(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-per.C:
			totalBytes := s.totalBytes.Load()
			s.bitrate.Store(8 * (totalBytes - prevBytes))
			prevBytes = totalBytes
		}
	}
}

func (s *Stats) update(unit units.Unit) {
	s.totalBytes.Add(uint32(len(unit.Payload)))
}

func (s *Stats) GetBitrate() uint32 {
	return s.bitrate.Load()
}

func (s *Stats) GetTotalBytes() uint32 {
	return s.totalBytes.Load()
}

//totalBytes := stats.GetTotalBytes()
//bps := 8 * (totalBytes - prevBytes)
//prevBytes = totalBytes
//_ = bps
