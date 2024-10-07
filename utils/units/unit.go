package units

import (
	"fmt"
)

type Unit struct {
	Payload   []byte
	PTS       int64
	DTS       int64
	Duration  int64
	TimeBase  int // audio: sampleRate, video: timebase
	Marker    bool
	FrameInfo FrameInfo
}

type FrameInfo struct {
	Flag          int
	PayloadHeader interface{}
}

func (u Unit) String() string {
	return fmt.Sprintf("Unit{PTS: %d, DTS: %d, Duration: %d, TimeBase: %d,  Payload: %d bytes, Marker:%v}",
		u.PTS, u.DTS, u.Duration, u.TimeBase, len(u.Payload), u.Marker)
}
