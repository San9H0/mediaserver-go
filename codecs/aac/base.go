package aac

import (
	"errors"
	"github.com/pion/rtp"
	"mediaserver-go/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Base struct {
}

func (b Base) MimeType() string {
	return "audio/aac"
}

func (b Base) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (b Base) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_AUDIO
}

func (b Base) CodecType() types.CodecType {
	return types.CodecTypeAAC
}

func (b Base) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_AAC
}

func (b Base) Extension() string {
	return "mp4"
}

func (b Base) RTPParser(cb func(codec codecs.Codec)) (codecs.RTPParser, error) {
	return nil, errors.New("aac codec not support rtp parser")
}

func (b Base) RTPIngressCapability() {

}

func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return nil, errors.New("aac codec not support rtp packetizer")
}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	config := NewConfig(Parameters{
		SampleRate:   param.SampleRate(),
		Channels:     param.Channels(),
		SampleFormat: param.Format(),
	})
	return NewAAC(config), nil
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}

func (b Base) GetBitStreamFilter(fromTranscoding bool) codecs.BitStreamFilter {
	return &BitStreamEmpty{}
}
