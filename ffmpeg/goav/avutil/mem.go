package avutil

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
//#include <stdlib.h>
import "C"
import "unsafe"

func AvFreep(buf *byte) {
	C.av_freep(unsafe.Pointer(&buf))
}
