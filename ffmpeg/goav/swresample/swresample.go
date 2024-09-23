package swresample

import "C"

/*
#cgo pkg-config: libswresample
#include <libswresample/swresample.h>
int myconvert(struct SwrContext *swr_ctx, int decoderSampleRate, int decoderNbSamples, int encoderSampleRate, int encoderNbChannels, int encoderSampleFmt, AVFrame* frame) {
	printf("myconvert called\n");
	int get_delay = swr_get_delay(swr_ctx, decoderSampleRate);
	printf("av_rescale_rnd called\n");
	int out_samples = av_rescale_rnd(get_delay + decoderNbSamples,
		encoderSampleRate,
		decoderSampleRate,
		AV_ROUND_UP);

	printf("av_samples_alloc_array_and_samples called\n");
	uint8_t** converted_data = NULL;
	int sample_size = av_samples_alloc_array_and_samples(&converted_data,
		NULL,
		encoderNbChannels,
		out_samples,
		encoderSampleFmt,
		0);
	printf("swr_convert called\n");
	int converted_sample_count = swr_convert(swr_ctx, converted_data, out_samples,(const uint8_t**)frame->data, frame->nb_samples);
	if (converted_sample_count < 0) {
		fprintf(stderr, "Error resampling frame.\n");
		return -1;
	}
	printf("myconvert ens\n");
	return 0;
}
*/
import "C"
import (
	"mediaserver-go/ffmpeg/goav/avutil"
	"unsafe"
)

type (
	SwrContext C.struct_SwrContext
)

func SwrAlloc() *SwrContext {
	return (*SwrContext)(C.swr_alloc())
}

func SwrAllocSetOpt2(
	outChLayout *avutil.AvChannelLayout,
	outSampleFmt avutil.AvSampleFormat,
	outSampleRate int,
	inChLayout *avutil.AvChannelLayout,
	inSampleFmt avutil.AvSampleFormat,
	inSampleRate int) *SwrContext {
	var swrCtx *SwrContext
	ret := int(C.swr_alloc_set_opts2((**C.struct_SwrContext)(unsafe.Pointer(&swrCtx)),
		(*C.struct_AVChannelLayout)(unsafe.Pointer(outChLayout)),
		(C.enum_AVSampleFormat)(outSampleFmt),
		C.int(outSampleRate),
		(*C.struct_AVChannelLayout)(unsafe.Pointer(inChLayout)),
		(C.enum_AVSampleFormat)(inSampleFmt),
		C.int(inSampleRate), 0, nil))
	if ret < 0 {
		return nil
	}
	return swrCtx
}

func (sc *SwrContext) SwrInit() int {
	return int(C.swr_init((*C.struct_SwrContext)(sc)))
}

// av_samples_alloc_array_and_samples(uint8_t ***audio_data, int *linesize, int nb_channels,
// int nb_samples, enum AVSampleFormat sample_fmt, int align);
//func (sc *SwrContext) AvSamplesAllocArrayAndSamples(sampleDelta, compensationDistance int) int {
//	return int(C.av_samples_alloc_array_and_samples((*C.struct_SwrContext)(sc), C.int(sampleDelta), C.int(compensationDistance)))
//}

func (sc *SwrContext) SwrConvert(out **uint8, outCount int, in **uint8, inCount int) int {
	return int(C.swr_convert((*C.struct_SwrContext)(unsafe.Pointer(sc)), (**C.uint8_t)(unsafe.Pointer(out)), C.int(outCount), (**C.uint8_t)(unsafe.Pointer(in)), C.int(inCount)))
}

func (sc *SwrContext) SwrConvert2(decoderSampleRate int, decoderNbSamples int, encoderSampleRate int, encoderNbChannels int, encoderSampleFmt int, frame *avutil.Frame) int {
	return int(C.myconvert((*C.struct_SwrContext)(unsafe.Pointer(sc)), C.int(decoderSampleRate), C.int(decoderNbSamples), C.int(encoderSampleRate), C.int(encoderNbChannels), C.int(encoderSampleFmt), (*C.struct_AVFrame)(unsafe.Pointer(frame))))
	//return int(C.swr_convert((*C.struct_SwrContext)(unsafe.Pointer(sc)), (**C.uint8_t)(unsafe.Pointer(out)), C.int(outCount), (**C.uint8_t)(unsafe.Pointer(in)), C.int(inCount)))
}

//swr_convert(struct SwrContext *s, uint8_t * const *out, int out_count,
//const uint8_t * const *in , int in_count);

//av_samples_alloc_array_and_samples(&converted_data, NULL, pCodecCtx->ch_layout.nb_channels, pFrame->nb_samples, AV_SAMPLE_FMT_S16, 0);

func (sc *SwrContext) GetDelay(sampleRate int) int64 {
	return int64(C.swr_get_delay((*C.struct_SwrContext)(unsafe.Pointer(sc)), C.int64_t(sampleRate)))
}

func SwrFree(sc *SwrContext) {
	C.swr_free((**C.struct_SwrContext)(unsafe.Pointer(&sc)))
}
