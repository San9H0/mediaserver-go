package rtpinbounder

import (
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"sync"
	"sync/atomic"
	"time"
)

const maxSN = 1 << 16

type Stats struct {
	mu sync.RWMutex

	ClockRate uint32
	SSRC      uint32

	// for stats
	totalBytes  uint32
	packetCount uint32
	packetLost  uint32
	maxSeqNo    uint16
	baseSeqNo   uint16
	cycle       uint32

	// for receiver report
	prevExpect     atomic.Uint32
	prevPacketLost atomic.Uint32
	lastSRRTPTime  atomic.Uint32
	lastSRNTPTime  atomic.Uint64
	lastSRTime     atomic.Int64
	jitter         float64
	lastTransit    uint32
}

func NewStats(clockRate, ssrc uint32) *Stats {
	return &Stats{
		ClockRate: clockRate,
		SSRC:      ssrc,
	}
}

func (s *Stats) CalcRTPStats(pkt *rtp.Packet, n int) {
	arrivalTime := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	sn := pkt.SequenceNumber
	if s.packetCount == 0 {
		s.baseSeqNo = sn
		s.maxSeqNo = sn
	} else if (sn-s.maxSeqNo)&0x8000 == 0 {
		if sn < s.maxSeqNo {
			s.cycle += maxSN
		}
		s.maxSeqNo = sn
	} else if (sn-s.maxSeqNo)&0x8000 > 0 {
		// 재전송필요
	}
	s.packetCount++
	s.totalBytes += uint32(n)

	arrival := uint32(arrivalTime / 1e6 * int64(s.ClockRate/1e3))
	transit := arrival - pkt.Timestamp
	if s.lastTransit != 0 {
		d := int32(transit - s.lastTransit)
		if d < 0 {
			d = -d
		}
		s.jitter += (float64(d) - s.jitter) / 16
	}
	s.lastTransit = transit
}

func (s *Stats) UpdateSR(rtcpPacket *rtcp.SenderReport) {
	s.lastSRRTPTime.Store(rtcpPacket.RTPTime)
	s.lastSRNTPTime.Store(rtcpPacket.NTPTime)
	s.lastSRTime.Store(time.Now().UnixNano())
}

func (s *Stats) extMaxSeqNo() uint32 {
	return s.cycle | uint32(s.maxSeqNo)
}

func (s *Stats) GetTotalBytes() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalBytes
}

func (s *Stats) GetReceiverReport() *rtcp.ReceiverReport {
	s.mu.RLock()
	ssrc := s.SSRC
	jitter := s.jitter
	maxSeqNo := s.maxSeqNo
	packetExpect := s.extMaxSeqNo() - uint32(s.baseSeqNo) + 1
	packetLost := packetExpect - s.packetCount
	s.mu.RUnlock()

	expectedInterval := packetExpect - s.prevExpect.Load()
	lostInterval := packetLost - s.prevPacketLost.Load()
	lostRate := float32(lostInterval) / float32(expectedInterval)
	fractionLost := uint8(lostRate * 256.0)
	lastSenderReport := uint32(s.lastSRNTPTime.Load() >> 16)

	var dlsr uint32
	lastSRtime := s.lastSRTime.Load()
	if lastSRtime != 0 {
		delayMS := uint32((time.Now().Nanosecond() - int(lastSRtime)) / 1e6)
		dlsr = (delayMS / 1e3) << 16
		dlsr |= (delayMS % 1e3) * 65536 / 1000
	}

	s.prevExpect.Store(packetExpect)
	s.prevPacketLost.Store(packetLost)

	return &rtcp.ReceiverReport{
		SSRC: ssrc,
		Reports: []rtcp.ReceptionReport{
			{
				SSRC:               ssrc,
				FractionLost:       fractionLost,
				TotalLost:          packetLost,
				LastSequenceNumber: uint32(maxSeqNo),
				Jitter:             uint32(jitter),
				LastSenderReport:   lastSenderReport,
				Delay:              dlsr,
			},
		},
	}
}

func (s *Stats) GetRemB() *rtcp.ReceiverEstimatedMaximumBitrate {
	return &rtcp.ReceiverEstimatedMaximumBitrate{
		Bitrate: float32(3_000_000),
		SSRCs:   []uint32{s.SSRC},
	}
}
