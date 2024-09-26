package opus

import (
	"errors"
	"github.com/pion/rtp"
	pinocodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Type struct {
}

func (t Type) MimeType() string {
	return pion.MimeTypeOpus
}

func (t Type) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (t Type) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_AUDIO
}

func (t Type) CodecType() types.CodecType {
	return types.CodecTypeOpus
}

func (t Type) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_OPUS
}

func (t Type) RTPParser(cb func([][]byte) [][]byte) (codecs.RTPParser, error) {
	return NewOpusParser(cb), nil
}

func (t Type) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pinocodecs.OpusPayloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

//func (t Type) NewConfig() codecs.Config {
//	return NewOpusConfig(Parameters{
//		Channels:   2,
//		SampleRate: 48000,
//		SampleFmt:  int(avutil.AV_SAMPLE_FMT_FLT),
//	})
//}

func (t Type) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return nil, errors.New("not supported until")
}
