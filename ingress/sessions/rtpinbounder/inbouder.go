package rtpinbounder

import (
	"context"
	"github.com/pion/rtp"
	"go.uber.org/zap"
	"mediaserver-go/hubs"
	"mediaserver-go/parsers"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type TrackContext struct {
	Timebase int
	Track    *hubs.HubSource
	Stats    *Stats
}

type Inbounder struct {
	ReadFunc func([]byte) (int, error)
	timebase int
	parser   parsers.RTPParser
}

func NewInbounder(parser parsers.RTPParser, timebase int, readFunc func([]byte) (int, error)) *Inbounder {
	return &Inbounder{
		parser:   parser,
		ReadFunc: readFunc,
		timebase: timebase,
	}
}

func (i *Inbounder) Run(ctx context.Context, hubTrack *hubs.HubSource, stats *Stats) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		buf := make([]byte, types.ReadBufferSize)

		n, err := i.ReadFunc(buf)
		if err != nil {
			return err
		}

		rtpPacket := &rtp.Packet{}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			log.Logger.Error("rtp failed to unmarshal", zap.Error(err))
			continue
		}

		if startTS == 0 {
			startTS = rtpPacket.Timestamp
		}
		pts := rtpPacket.Timestamp - startTS

		if rtpPacket.Timestamp != prevTS {
			if prevTS == 0 {
				duration = 0
			} else {
				duration = int(rtpPacket.Timestamp - prevTS)
			}
		}

		stats.CalcRTPStats(rtpPacket, n)

		for _, payload := range i.parser.Parse(rtpPacket) {
			hubTrack.Write(units.Unit{
				Payload:  payload,
				PTS:      int64(pts),
				DTS:      int64(pts),
				Duration: int64(duration),
				TimeBase: i.timebase,
			})
		}
		prevTS = rtpPacket.Timestamp
	}

}
