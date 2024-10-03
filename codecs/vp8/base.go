package vp8

import (
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
	return pion.MimeTypeVP8
}

func (b Base) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (b Base) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_VIDEO
}

func (b Base) CodecType() types.CodecType {
	return types.CodecTypeVP8
}

func (b Base) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_VP8
}

func (b Base) Extension() string {
	return "mp4"
}

func (b Base) RTPParser(cb func(codec codecs.Codec)) (codecs.RTPParser, error) {
	return NewRTPParser(cb), nil
}

func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.VP8Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	config := Config{}
	config.Width = param.Width()
	config.Height = param.Height()
	vp8Codec := NewVP8(&config)
	return vp8Codec, nil
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}

func (b Base) GetBitStreamFilter(fromTranscoding bool) codecs.BitStreamFilter {
	return &BitStreamEmpty{}
}
