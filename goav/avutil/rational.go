package avutil

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
import "C"

func (r Rational) Num() int {
	return int(r.num)
}

func (r Rational) SetNum(num int) {
	r.num = C.int(num)
}

func (r Rational) Den() int {
	return int(r.den)
}

func (r Rational) SetDen(den int) {
	r.den = C.int(den)
}

func NewRational(num, den int) Rational {
	return Rational{
		num: C.int(num),
		den: C.int(den),
	}
}

func GetDurationMicroSec(r Rational, duration int64) int {
	return 1000 * 1000 * r.Num() / r.Den() * int(duration)
}

func GetTimebaseUSec(r Rational, ts int64) int64 {
	return 1000 * 1000 * ts * int64(r.Num()) / int64(r.Den())
}
