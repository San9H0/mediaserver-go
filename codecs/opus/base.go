package opus

import (
	"github.com/pion/rtp"
	pinocodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Base struct {
}

func (b Base) MimeType() string {
	return pion.MimeTypeOpus
}

func (b Base) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (b Base) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_AUDIO
}

func (b Base) CodecType() types.CodecType {
	return types.CodecTypeOpus
}

func (b Base) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_OPUS
}

func (b Base) Extension() string {
	return "mp4"
}

func (b Base) RTPParser(cb func(codec codecs.Codec)) (codecs.RTPParser, error) {
	return NewRTPParser(cb), nil
}

func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pinocodecs.OpusPayloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

//func (t Type) NewConfig() codecs.Config {
//	return NewOpusConfig(Parameters{
//		Channels:   2,
//		SampleRate: 48000,
//		SampleFmt:  int(avutil.AV_SAMPLE_FMT_FLT),
//	})
//}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	config := NewConfig(Parameters{
		Channels:     2,
		SampleRate:   48000,
		SampleFormat: int(avutil.AV_SAMPLE_FMT_FLT),
	})
	return NewOpus(config), nil
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}

func (b Base) GetBitStreamFilter(fromTranscoding bool) codecs.BitStreamFilter {
	return &BitStreamEmpty{}
}
