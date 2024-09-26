package avutil

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
//#include <libavutil/channel_layout.h>
import "C"
import "unsafe"

type (
	AvChannelLayout C.struct_AVChannelLayout
)

func AvChannelLayoutDefault(chLayout *AvChannelLayout, nbChannels int) {
	C.av_channel_layout_default((*C.struct_AVChannelLayout)(unsafe.Pointer(chLayout)), C.int(nbChannels))
}

//func ToCAvChannelLayout(chLayout AvChannelLayout) C.struct_AVChannelLayout {
//	return C.struct_AVPacket(chLayout)
//}

func FromCAvChannelLayout(chLayout C.struct_AVChannelLayout) AvChannelLayout {
	return (AvChannelLayout)(chLayout)
}

func (a *AvChannelLayout) NbChannels() int {
	return int(a.nb_channels)
}
