package transcoders

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/thirdparty/ffmpeg/swresample"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
	"unsafe"
)

type AudioTranscoder struct {
	source, target codecs.Codec

	decoder    *avcodec.Codec
	decoderCtx *avcodec.CodecContext
	swrCtx     *swresample.SwrContext

	encoder    *avcodec.Codec
	encoderCtx *avcodec.CodecContext

	audioFifo *avutil.AvAudioFifo

	read int64

	reuseBuffer **uint8
}

func NewAudioTranscoder(source, target codecs.Codec) *AudioTranscoder {
	return &AudioTranscoder{
		source: source,
		target: target,
	}
}

func (t *AudioTranscoder) Source() codecs.Codec {
	return t.source
}

func (t *AudioTranscoder) Target() codecs.Codec {
	return t.target
}

func (t *AudioTranscoder) Close() {
	//if t.decoderCtx != nil {
	//	avcodec.AvCodecFreeContext(&t.decoderCtx)
	//}
	//if t.encoderCtx != nil {
	//	avcodec.AvCodecFreeContext(&t.encoderCtx)
	//}
	//if t.swrCtx != nil {
	//	swresample.SwrFree(t.swrCtx)
	//}
	//if t.audioFifo != nil {
	//	avutil.AvAudioFifoFree(t.audioFifo)
	//}
	//if t.reuseBuffer != nil {
	//	avutil.AvFreep(t.reuseBuffer)
	//}
}

func (t *AudioTranscoder) Setup() error {
	source, ok := t.source.(codecs.AudioCodec)
	if !ok {
		return errors.New("invalid source codec")
	}
	target, ok := t.target.(codecs.AudioCodec)
	if !ok {
		return errors.New("invalid target codec")
	}
	decoder := avcodec.AvcodecFindDecoder(source.AVCodecID())

	decoderCtx := decoder.AvCodecAllocContext3()
	if decoderCtx == nil {
		return errors.New("avcodec alloc context3 failed")
	}

	source.SetCodecContext(decoderCtx, nil)
	if decoderCtx.AvCodecOpen2(decoder, nil) < 0 {
		return errors.New("avcodec open failed")
	}

	encoder := avcodec.AvcodecFindEncoder(target.AVCodecID())
	encoderCtx := encoder.AvCodecAllocContext3()
	target.SetCodecContext(encoderCtx, nil)
	if encoderCtx.AvCodecOpen2(encoder, nil) < 0 {
		return errors.New("avcodec open failed")
	}

	swrCtx := swresample.SwrAllocSetOpt2(
		encoderCtx.ChLayout(), encoderCtx.SampleFmt(), encoderCtx.SampleRate(),
		decoderCtx.ChLayout(), decoderCtx.SampleFmt(), decoderCtx.SampleRate())
	if swrCtx.SwrInit() < 0 {
		return errors.New("swr init failed")
	}

	audioFifo := target.AvCodecFifoAlloc()

	t.decoder = decoder
	t.decoderCtx = decoderCtx
	t.encoder = encoder
	t.encoderCtx = encoderCtx
	t.swrCtx = swrCtx
	t.audioFifo = audioFifo

	var pbuffer = (**uint8)(unsafe.Pointer(nil))
	if avutil.AvSamplesAllocArrayAndSamples(&pbuffer, t.encoderCtx.ChLayout().NbChannels(), 2000, t.encoderCtx.SampleFmt()) <= 0 {
		fmt.Println("av_samples_alloc_array_and_samples failed")
		return nil
	}
	t.reuseBuffer = pbuffer
	return nil
}

func (t *AudioTranscoder) Transcode(unit units.Unit) []units.Unit {
	// TODO

	var result []units.Unit
	pkt := avcodec.AvPacketAlloc()
	pkt.SetPTS(unit.PTS)
	pkt.SetDTS(unit.DTS)
	pkt.SetData(unit.Payload)
	pkt.SetDuration(unit.Duration)

	if ret := t.decoderCtx.AvCodecSendPacket(pkt); ret < 0 {
		log.Logger.Error("AvCodecSendPacket failed", zap.Error(errors.New(avutil.AvErr2str(ret))))
		return nil
	}

	frame := avutil.AvFrameAlloc()
	if ret := t.decoderCtx.AvCodecReceiveFrame(frame); ret < 0 {
		log.Logger.Error("AvCodecReceiveFrame failed", zap.Error(errors.New(avutil.AvErr2str(ret))))
		return nil
	}

	delay := t.swrCtx.GetDelay(t.decoderCtx.SampleRate())
	outSamples := avutil.AvRescaleRnd(int(delay)+frame.NbSamples(), t.encoderCtx.SampleRate(), t.decoderCtx.SampleRate(), avutil.AV_ROUND_UP)
	sampleCount := t.swrCtx.SwrConvert(t.reuseBuffer, outSamples, frame.GetDataP(), frame.NbSamples())
	if sampleCount < 0 {
		fmt.Println("swr_convert failed")
		return nil
	}

	written := t.audioFifo.AvAudioFifoWrite(t.reuseBuffer, sampleCount)
	if written < 0 {
		fmt.Println("av_audio_fifo_write failed")
		return nil
	}

	for t.audioFifo.AvAudioFifoSize() >= t.encoderCtx.FrameSize() {
		opusFrame := avutil.AvFrameAlloc()
		opusFrame.SetNbSamples(t.encoderCtx.FrameSize())
		avutil.AvChannelLayoutDefault(opusFrame.ChLayout(), 2)
		opusFrame.SetFormat(int(t.encoderCtx.SampleFmt()))
		opusFrame.SetSampleRate(t.encoderCtx.SampleRate())
		opusFrame.AvFrameGetBuffer(0)
		read := t.audioFifo.AvAudioFifoRead(opusFrame.GetDataP(), t.encoderCtx.FrameSize())
		pts := t.read
		t.read += int64(read)
		opusFrame.SetPTS(pts)
		opusFrame.SetDTS(pts)

		if t.encoderCtx.AvCodecSendFrame(opusFrame) < 0 {
			fmt.Println("AvCodecSendFrame failed")
			return nil
		}

		recvPkt := avcodec.AvPacketAlloc()
		if ret := t.encoderCtx.AvCodecReceivePacket(recvPkt); ret < 0 {
			fmt.Println("AvCodecReceivePacket failed")
			return nil
		}

		result = append(result, units.Unit{
			Payload:  recvPkt.Data(),
			PTS:      recvPkt.PTS(),
			DTS:      recvPkt.DTS(),
			Duration: recvPkt.Duration(),
			TimeBase: unit.TimeBase,
		})
	}

	return result
}
