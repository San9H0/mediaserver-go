package avutil

// 	#cgo pkg-config: libavutil
// 	#include <libavutil/avutil.h>
// 	#include <libavutil/frame.h>
import "C"
import (
	"unsafe"
)

func AvFrameAlloc() *Frame {
	return (*Frame)(C.av_frame_alloc())
}

func (f *Frame) GetDataP() **uint8 {
	return (**uint8)(unsafe.Pointer(&f.data[0]))
}

func (f *Frame) GetData() []byte {
	result := make([]byte, f.NbSamples()*f.ChLayout().NbChannels()*2)
	for i := 0; i < f.NbSamples(); i++ {
		for ch := 0; ch < f.ChLayout().NbChannels(); ch++ {
			result[ch+i*2+0] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(f.GetDataP())) + uintptr(ch)*2 + uintptr(i)*2))
			result[ch+i*2+1] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(f.GetDataP())) + uintptr(ch)*2 + uintptr(i)*2 + 1))
		}
	}
	return result
}

func (f *Frame) NbSamples() int {
	return int(f.nb_samples)
}

func (f *Frame) SetNbSamples(nbSamples int) {
	f.nb_samples = C.int(nbSamples)
}

func (f *Frame) SampleRate() int {
	return int(f.sample_rate)
}

func (f *Frame) SetSampleRate(sampleRate int) {
	f.sample_rate = C.int(sampleRate)
}

func (f *Frame) Format() int {
	return int(f.format)
}

func (f *Frame) SetFormat(format int) {
	f.format = C.int(format)
}

func (f *Frame) PTS() int {
	return int(f.pts)
}

func (f *Frame) SetPTS(pts int64) {
	f.pts = C.int64_t(pts)
}

func (f *Frame) DTS() int {
	return int(f.pkt_dts)
}

func (f *Frame) SetDTS(dts int64) {
	f.pkt_dts = C.int64_t(dts)
}

func (f *Frame) ChLayout() *AvChannelLayout {
	return (*AvChannelLayout)(unsafe.Pointer(&f.ch_layout))
}

func (f *Frame) AvFrameGetBuffer(n int) int {
	return int(C.av_frame_get_buffer((*C.struct_AVFrame)(unsafe.Pointer(f)), C.int(n)))
}
