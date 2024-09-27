package transcoders

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/bitstreamfilter"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/thirdparty/ffmpeg/swresample"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
)

var (
	errFailedToSetTranscodeCodec = errors.New("failed to set transcode codec")
)

type VideoTranscoder struct {
	decoderBitStreamFilter bitstreamfilter.BitStreamFilter
	decoder                *avcodec.Codec
	decoderCtx             *avcodec.CodecContext
	swrCtx                 *swresample.SwrContext

	encoder    *avcodec.Codec
	encoderCtx *avcodec.CodecContext
}

func NewVideoTranscoder() *VideoTranscoder {
	return &VideoTranscoder{}
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

func (t *VideoTranscoder) Setup(source, target codecs.Codec) error {
	bitStreamFilter := source.GetBitStreamFilter()
	decoder := avcodec.AvcodecFindDecoder(source.AVCodecID())
	if decoder == nil {
		return fmt.Errorf("could not find decoder: %w", errFailedToSetTranscodeCodec)
	}
	decoderCtx := decoder.AvCodecAllocContext3()
	if decoderCtx == nil {
		return fmt.Errorf("could not allocate codec context: %w", errFailedToSetTranscodeCodec)
	}
	source.SetCodecContext(decoderCtx)
	if decoderCtx.AvCodecOpen2(decoder, nil) < 0 {
		return fmt.Errorf("could not open codec: %w", errFailedToSetTranscodeCodec)
	}

	encoder := avcodec.AvcodecFindEncoder(target.AVCodecID())
	if encoder == nil {
		return fmt.Errorf("could not find encoder: %w", errFailedToSetTranscodeCodec)
	}
	encoderCtx := encoder.AvCodecAllocContext3()
	if encoderCtx == nil {
		return fmt.Errorf("could not allocate codec context: %w", errFailedToSetTranscodeCodec)
	}
	target.SetCodecContext(encoderCtx)
	if encoderCtx.AvCodecOpen2(encoder, nil) < 0 {
		return fmt.Errorf("could not open codec: %w", errFailedToSetTranscodeCodec)
	}

	t.decoderBitStreamFilter = bitStreamFilter
	t.decoder = decoder
	t.decoderCtx = decoderCtx
	t.encoder = encoder
	t.encoderCtx = encoderCtx

	return nil
}

func (t *VideoTranscoder) Transcode(unit units.Unit) []units.Unit {
	payload := t.decoderBitStreamFilter.AddFilter(unit)

	pkt := avcodec.AvPacketAlloc()
	pkt.SetPTS(unit.PTS)
	pkt.SetDTS(unit.DTS)
	pkt.SetDuration(unit.Duration)
	pkt.SetData(payload)

	if ret := t.decoderCtx.AvCodecSendPacket(pkt); ret < 0 {
		log.Logger.Error("AvCodecSendPacket failed", zap.Error(errors.New(avutil.AvErr2str(ret))))
		return nil
	}

	frame := avutil.AvFrameAlloc()
	if ret := t.decoderCtx.AvCodecReceiveFrame(frame); ret < 0 {
		fmt.Println("ret:", ret)
		if avutil.AvAgain(ret) {
			return nil
		}
		log.Logger.Error("AvCodecReceiveFrame failed", zap.Error(errors.New(avutil.AvErr2str(ret))))
		return nil
	}

	//frame.SetSampleRate(t.encoderCtx.SampleRate())
	//frame.SetFormat(int(t.encoderCtx.PixelFormat()))
	//frame.SetNbSamples(t.encoderCtx.FrameSize())
	frame.SetPictType(avutil.AV_PICTURE_TYPE_NONE)

	if t.encoderCtx.AvCodecSendFrame(frame) < 0 {
		fmt.Println("AvCodecSendFrame failed")
		return nil
	}

	frame.AvFrameFree()

	recvPkt := avcodec.AvPacketAlloc()
	if ret := t.encoderCtx.AvCodecReceivePacket(recvPkt); ret < 0 {
		if avutil.AvAgain(ret) {
			return nil
		}
		log.Logger.Error("AvCodecReceivePacket failed", zap.Error(errors.New(avutil.AvErr2str(ret))))
		return nil
	}

	var result []units.Unit

	result = append(result, units.Unit{
		Payload:  recvPkt.Data()[4:],
		PTS:      recvPkt.PTS(),
		DTS:      recvPkt.DTS(),
		Duration: recvPkt.Duration(),
		TimeBase: t.encoderCtx.SampleRate(),
	})

	recvPkt.AvPacketFree()

	return result
}
