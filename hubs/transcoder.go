package hubs

import (
	"errors"
	"fmt"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avformat"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type AudioTranscoder struct {
	outputFmtCtx *avformat.FormatContext

	decoder    *avcodec.Codec
	decoderCtx *avcodec.CodecContext

	encoder    *avcodec.Codec
	encoderCtx *avcodec.CodecContext

	audioFifo *avutil.AvAudioFifo
}

func NewAudioTranscoder() *AudioTranscoder {
	return &AudioTranscoder{}
}

func (t *AudioTranscoder) SetupAudio(sourceCodec, targetCodec codecs.Codec) error {
	source, ok := sourceCodec.(codecs.AudioCodec)
	if !ok {
		return errors.New("invalid source codec")
	}
	target, ok := targetCodec.(codecs.AudioCodec)
	if !ok {
		return errors.New("invalid target codec")
	}
	decoder := avcodec.AvcodecFindDecoder(types.CodecIDFromType(source.CodecType()))
	decoderCtx := decoder.AvCodecAllocContext3()
	source.SetCodecContext(decoderCtx)
	if decoderCtx.AvCodecOpen2(decoder, nil) < 0 {
		return errors.New("avcodec open failed")
	}

	t.decoder = decoder
	t.decoderCtx = decoderCtx

	encoder := avcodec.AvcodecFindEncoder(types.CodecIDFromType(target.CodecType()))
	encoderCtx := encoder.AvCodecAllocContext3()
	target.SetCodecContext(encoderCtx)
	if encoderCtx.AvCodecOpen2(encoder, nil) < 0 {
		return errors.New("avcodec open failed")
	}
	t.encoder = encoder
	t.encoderCtx = encoderCtx

	outputFmtCtx := avformat.NewAvFormatContextNull()
	if avformat.AvformatAllocOutputContext2(&outputFmtCtx, nil, "", fmt.Sprintf("output.%s", "opus")) < 0 {
		return errors.New("avformat context allocation failed")
	}
	stream := outputFmtCtx.AvformatNewStream(nil)
	if stream == nil {
		return errors.New("avformat new stream failed")
	}

	if avcodec.AvCodecParametersFromContext(stream.CodecParameters(), encoderCtx) < 0 {
		return errors.New("avcodec parameters from context failed")
	}

	outputFmtCtx.SetPb(avformat.AVIOOpenDynBuf())

	if outputFmtCtx.AvformatWriteHeader(nil) < 0 {
		return errors.New("avformat write header failed")
	}

	t.outputFmtCtx = outputFmtCtx
	audioFifo := target.AvCodecFifoAlloc()
	t.audioFifo = audioFifo

	return nil
}

func (t *AudioTranscoder) Transcode(unit units.Unit) []units.Unit {
	pkt := avcodec.AvPacketAlloc()
	pkt.SetPTS(unit.PTS)
	pkt.SetDTS(unit.DTS)
	pkt.SetData(unit.Payload)
	pkt.SetDuration(unit.Duration)
	if ret := t.decoderCtx.AvCodecSendPacket(pkt); ret < 0 {
		fmt.Println("err:", ret, ", avutil.AvErr2str(ret):", avutil.AvErr2str(ret))
		return nil
	}

	frame := avutil.AvFrameAlloc()
	if ret := t.decoderCtx.AvCodecReceiveFrame(frame); ret < 0 {
		fmt.Println("err:", ret, ", avutil.AvErr2str(ret):", avutil.AvErr2str(ret))
		return nil
	}

	written := t.audioFifo.AvAudioFifoWrite(frame.GetDataP(), frame.NbSamples())
	if written < 0 {
		fmt.Println("av_audio_fifo_write failed")
		return nil
	}

	nbSamples := 0

	var result []units.Unit
	for t.audioFifo.AvAudioFifoSize() >= t.encoderCtx.FrameSize() {
		frame := avutil.AvFrameAlloc()
		frame.SetNbSamples(t.encoderCtx.FrameSize())
		avutil.AvChannelLayoutDefault(frame.ChLayout(), 2)
		frame.SetFormat(int(t.encoderCtx.SampleFmt()))
		frame.SetSampleRate(t.encoderCtx.SampleRate())
		frame.SetPTS(unit.PTS)
		frame.SetDTS(unit.DTS)
		nbSamples += frame.NbSamples()
		frame.AvFrameGetBuffer(0)
		read := t.audioFifo.AvAudioFifoRead(frame.GetDataP(), t.encoderCtx.FrameSize())
		fmt.Println(" read:", read)

		if t.encoderCtx.AvCodecSendFrame(frame) < 0 {
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
			TimeBase: t.encoderCtx.SampleRate(),
		})
	}

	return result
}
