package h264

import (
	"fmt"
	"github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/parsers/format"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Base struct {
}

func (b Base) MimeType() string {
	return pion.MimeTypeH264
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
	return avcodec.AV_CODEC_ID_H264
}

// RTPParser 는 RTP Packets들을 코덱의 고유 유닛으로 파싱하거나, 코덱의 고유 유닛을 RTP 패킷으로 패킷화 한다.
// cb: 비트 스트림중 codec정보를 읽으면 콜백을 발생한다. 비디오의 정보가 변경되는 경우 호출됨.
func (b Base) RTPParser(cb func(codecs.Codec)) (codecs.RTPParser, error) {
	return NewH264Parser(cb), nil
}

// RTPPacketizer 는 RTP Packets들을 코덱의 고유 유닛으로 파싱하거나, 코덱의 고유 유닛을 RTP 패킷으로 패킷화 한다.
func (b Base) RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error) {
	return rtp.NewPacketizer(types.MTUSize, pt, ssrc, &pioncodecs.H264Payloader{}, rtp.NewRandomSequencer(), clockRate), nil
}

func (b Base) CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (codecs.Codec, error) {
	sps, pps := format.SPSPPSFromAVCCExtraData(param.ExtraData())
	config := &Config{}
	err := config.UnmarshalFromSPSPPS(sps, pps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sps pps: %v", err)
	}
	h264Codecs := NewH264(config)
	return h264Codecs, nil
}

func (b Base) Decoder() codecs.Decoder {
	return &Decoder{}
}
