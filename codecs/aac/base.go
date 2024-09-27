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

func (t Base) MimeType() string {
	return "audio/aac"
}

func (t Base) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (t Base) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_AUDIO
}

func (t Base) CodecType() types.CodecType {
	return types.CodecTypeAV1
}

func (t Base) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_AAC
}

func (t Base) RTPParser(cb func(codec codecs.Codec)) (codecs.RTPParser, error) {
	return nil, errors.New("aac codec not support rtp parser")
}

func (t Base) RTPIngressCapability() {

}

func (t Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return nil, errors.New("aac codec not support rtp packetizer")
}

func (t Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
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
