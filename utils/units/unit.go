package units

import "fmt"

type Unit struct {
	Payload  []byte
	PTS      int64
	DTS      int64
	Duration int64
	TimeBase int // audio: sampleRate, video: timebase

	Flags int
}

func (u Unit) String() string {
	return fmt.Sprintf("Unit{PTS: %d, DTS: %d, Duration: %d, TimeBase: %d, Flags: %d, Payload: %d bytes}",
		u.PTS, u.DTS, u.Duration, u.TimeBase, u.Flags, len(u.Payload))
}
