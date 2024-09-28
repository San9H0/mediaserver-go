package vp8

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
	//var codec *VP8
	//return NewRTPParser(func(datas [][]byte) [][]byte {
	//	for _, d := range datas {
	//		header, ok := GetFrameHeader(d)
	//		if ok {
	//			if codec == nil || (codec.Width() != header.Width || codec.Height() != header.Height) {
	//				codec = NewVP8(header.Width, header.Height)
	//			}
	//			break
	//		}
	//	}
	//	return datas
	//}), nil
}

func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.VP8Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return nil, errors.New("not supported until")
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}
