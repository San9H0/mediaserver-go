package avutil

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
//#include <stdlib.h>
import "C"
import "unsafe"

func AvFreep(buf **uint8) {
	C.av_freep(unsafe.Pointer(buf))
	C.av_free(unsafe.Pointer(buf))
}
