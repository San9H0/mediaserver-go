package aac

import (
	"errors"
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Type struct {
}

func (t Type) MimeType() string {
	return "audio/aac"
}

func (t Type) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (t Type) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_AUDIO
}

func (t Type) CodecType() types.CodecType {
	return types.CodecTypeAV1
}

func (t Type) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_AAC
}

func (t Type) RTPParser() (codecs.RTPParser, error) {
	return nil, errors.New("aac codec not support rtp parser")
}

func (t Type) RTPIngressCapability() {

}

func (t Type) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return nil, errors.New("aac codec not support rtp packetizer")
}

func (t Type) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return NewAAC(Parameters{
		SampleRate: param.SampleRate(),
		Channels:   param.Channels(),
		SampleFmt:  param.Format(),
	}), nil
}
