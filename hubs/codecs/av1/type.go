package av1

import (
	"errors"
	"github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Type struct {
}

func (t Type) MimeType() string {
	return pion.MimeTypeAV1
}

func (t Type) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (t Type) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_VIDEO
}

func (t Type) CodecType() types.CodecType {
	return types.CodecTypeAV1
}

func (t Type) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_AV1
}

func (t Type) RTPParser(cb func([][]byte) [][]byte) (codecs.RTPParser, error) {
	return NewAV1Parser(cb), nil
}

func (t Type) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.AV1Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (t Type) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return nil, errors.New("not supported until")
}
