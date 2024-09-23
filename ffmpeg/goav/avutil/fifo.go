package avutil

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
//#include <libavutil/audio_fifo.h>
//#include <stdlib.h>
import "C"
import "unsafe"

func AvAudioFifoAlloc(sampleFmt AvSampleFormat, channels, nbSamples int) *AvAudioFifo {
	return (*AvAudioFifo)(C.av_audio_fifo_alloc(C.enum_AVSampleFormat(sampleFmt),
		C.int(channels), C.int(nbSamples)))
}

func (a *AvAudioFifo) AvAudioFifoSize() int {
	return int(C.av_audio_fifo_size((*C.struct_AVAudioFifo)(a)))
}

func (a *AvAudioFifo) AvAudioFifoWrite(data **uint8, nbSamples int) int {
	return int(C.av_audio_fifo_write((*C.struct_AVAudioFifo)(a),
		(*unsafe.Pointer)(unsafe.Pointer(data)),
		C.int(nbSamples)))
}

func (a *AvAudioFifo) AvAudioFifoRead(data **uint8, nbSamples int) int {
	return int(C.av_audio_fifo_read((*C.struct_AVAudioFifo)(a),
		(*unsafe.Pointer)(unsafe.Pointer(data)),
		C.int(nbSamples)))
}

func AvAudioFifoFree(a *AvAudioFifo) {
	C.av_audio_fifo_free((*C.struct_AVAudioFifo)(unsafe.Pointer(a)))
}
