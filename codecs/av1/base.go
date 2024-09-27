package av1

import (
	"errors"
	"github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Base struct {
}

func (b Base) MimeType() string {
	return pion.MimeTypeAV1
}

func (b Base) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (b Base) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_VIDEO
}

func (b Base) CodecType() types.CodecType {
	return types.CodecTypeAV1
}

func (b Base) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_AV1
}

func (b Base) RTPParser(cb func(codec codecs.Codec)) (codecs.RTPParser, error) {
	return NewRTPParser(cb), nil
}

func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.AV1Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return nil, errors.New("not supported until")
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}
