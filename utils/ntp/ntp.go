package ntp

import "time"

var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

type NTPTime uint64

func (t NTPTime) Duration() time.Duration {
	sec := (t >> 32) * 1e9
	frac := (t & 0xFFFFFFFF) * 1e9
	nsec := frac >> 32
	if uint32(frac) >= 0x80000000 {
		nsec++
	}
	return time.Duration(sec + nsec)
}

func (t NTPTime) Time() time.Time {
	return ntpEpoch.Add(t.Duration())
}

func GetNTPTime(t time.Time) NTPTime {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / 1e9
	nsec = (nsec - sec*1e9) << 32
	frac := nsec / 1e9
	if nsec%1e9 >= 1e9/2 {
		frac++
	}
	return NTPTime(sec<<32 | frac)
}
