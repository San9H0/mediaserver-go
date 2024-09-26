package vp8

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
	return pion.MimeTypeVP8
}

func (t Type) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (t Type) AVMediaType() avutil.MediaType {
	return avutil.AVMEDIA_TYPE_VIDEO
}

func (t Type) CodecType() types.CodecType {
	return types.CodecTypeVP8
}

func (t Type) AVCodecID() avcodec.CodecID {
	return avcodec.AV_CODEC_ID_VP8
}
func (t Type) RTPParser(cb func([][]byte) [][]byte) (codecs.RTPParser, error) {
	var codec *VP8
	return NewVP8Parser(func(datas [][]byte) [][]byte {
		for _, d := range datas {
			header, ok := GetFrameHeader(d)
			if ok {
				if codec == nil || (codec.Width() != header.Width || codec.Height() != header.Height) {
					codec = NewVP8(header.Width, header.Height)
				}
				break
			}
		}
		return datas
	}), nil
}

func (t Type) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.VP8Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (t Type) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	return nil, errors.New("not supported until")
}
