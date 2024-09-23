package avutil

//#include <libavutil/samplefmt.h>
import "C"
import "unsafe"

type (
	AvSampleFormat = C.enum_AVSampleFormat
)

const (
	AV_SAMPLE_FMT_NONE AvSampleFormat = C.AV_SAMPLE_FMT_NONE
	AV_SAMPLE_FMT_U8   AvSampleFormat = C.AV_SAMPLE_FMT_U8
	AV_SAMPLE_FMT_S16  AvSampleFormat = C.AV_SAMPLE_FMT_S16
	AV_SAMPLE_FMT_S32  AvSampleFormat = C.AV_SAMPLE_FMT_S32
	AV_SAMPLE_FMT_FLT  AvSampleFormat = C.AV_SAMPLE_FMT_FLT
	AV_SAMPLE_FMT_DBL  AvSampleFormat = C.AV_SAMPLE_FMT_DBL
	AV_SAMPLE_FMT_U8P  AvSampleFormat = C.AV_SAMPLE_FMT_U8P
	AV_SAMPLE_FMT_S16P AvSampleFormat = C.AV_SAMPLE_FMT_S16P
	AV_SAMPLE_FMT_S32P AvSampleFormat = C.AV_SAMPLE_FMT_S32P
	AV_SAMPLE_FMT_FLTP AvSampleFormat = C.AV_SAMPLE_FMT_FLTP
	AV_SAMPLE_FMT_DBLP AvSampleFormat = C.AV_SAMPLE_FMT_DBLP
	AV_SAMPLE_FMT_S64  AvSampleFormat = C.AV_SAMPLE_FMT_S64
	AV_SAMPLE_FMT_S64P AvSampleFormat = C.AV_SAMPLE_FMT_S64P
	AV_SAMPLE_FMT_NB   AvSampleFormat = C.AV_SAMPLE_FMT_NB
)

func AvSamplesAllocArrayAndSamples(pdata ***uint8, nbChannels int, nbSamples int, sampleFmt AvSampleFormat) int {
	//sampleSize := AvGetBytesPerSample(sampleFmt)
	//b := make([]byte, nbChannels*nbSamples*sampleSize)
	//return b
	return int(C.av_samples_alloc_array_and_samples((***C.uint8_t)(unsafe.Pointer(pdata)), nil, C.int(nbChannels), C.int(nbSamples), (C.enum_AVSampleFormat)(sampleFmt), 0))
}

func AvGetBytesPerSample(sampleFmt AvSampleFormat) int {
	return int(C.av_get_bytes_per_sample((C.enum_AVSampleFormat)(sampleFmt)))
}
