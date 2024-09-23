package transcoders

import (
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/swresample"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/units"
)

type VideoTranscoder struct {
	decoder    *avcodec.Codec
	decoderCtx *avcodec.CodecContext
	swrCtx     *swresample.SwrContext

	encoder    *avcodec.Codec
	encoderCtx *avcodec.CodecContext
}

func (t *VideoTranscoder) Close() {
	if t.decoderCtx != nil {
		avcodec.AvCodecFreeContext(&t.decoderCtx)
	}
	if t.encoderCtx != nil {
		avcodec.AvCodecFreeContext(&t.encoderCtx)
	}
	if t.swrCtx != nil {
		swresample.SwrFree(t.swrCtx)
	}
}

func (t *VideoTranscoder) Setup(sourceCodec, targetCodec codecs.Codec) error {
	//source, ok := sourceCodec.(codecs.VideoCodec)
	//if !ok {
	//	return errors.New("invalid source codec")
	//}
	//target, ok := targetCodec.(codecs.VideoCodec)
	//if !ok {
	//	return errors.New("invalid target codec")
	//}
	//decoder := avcodec.AvcodecFindDecoder(types.CodecIDFromType(source.CodecType()))
	//
	//decoderCtx := decoder.AvCodecAllocContext3()
	//if decoderCtx == nil {
	//	return errors.New("avcodec alloc context3 failed")
	//}
	//
	//source.SetCodecContext(decoderCtx)
	//if decoderCtx.AvCodecOpen2(decoder, nil) < 0 {
	//	return errors.New("avcodec open failed")
	//}
	//
	//encoder := avcodec.AvcodecFindEncoder(types.CodecIDFromType(target.CodecType()))
	//encoderCtx := encoder.AvCodecAllocContext3()
	//target.SetCodecContext(encoderCtx)
	//if encoderCtx.AvCodecOpen2(encoder, nil) < 0 {
	//	return errors.New("avcodec open failed")
	//}
	//
	//swrCtx := swresample.SwrAllocSetOpt2(
	//	encoderCtx.ChLayout(), encoderCtx.SampleFmt(), encoderCtx.SampleRate(),
	//	decoderCtx.ChLayout(), decoderCtx.SampleFmt(), decoderCtx.SampleRate())
	//if swrCtx.SwrInit() < 0 {
	//	return errors.New("swr init failed")
	//}
	//
	//audioFifo := target.AvCodecFifoAlloc()
	//
	//t.decoder = decoder
	//t.decoderCtx = decoderCtx
	//t.encoder = encoder
	//t.encoderCtx = encoderCtx
	//t.swrCtx = swrCtx
	//t.audioFifo = audioFifo
	//
	//var pbuffer = (**uint8)(unsafe.Pointer(nil))
	//if avutil.AvSamplesAllocArrayAndSamples(&pbuffer, t.encoderCtx.ChLayout().NbChannels(), 2000, t.encoderCtx.SampleFmt()) <= 0 {
	//	fmt.Println("av_samples_alloc_array_and_samples failed")
	//	return nil
	//}
	//t.reuseBuffer = pbuffer

	return nil
}

func (t *VideoTranscoder) Transcode(unit units.Unit) []units.Unit {
	return nil
}
