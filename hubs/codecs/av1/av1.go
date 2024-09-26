package av1

import (
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
	"strings"
)

type AV1 struct {
	width, height int
}

func NewAV1(width, height int) (*AV1, error) {
	return &AV1{
		width: width, height: height,
	}, nil
}

func (v *AV1) Clone() codecs.Codec {
	return &AV1{
		width: v.width, height: v.height,
	}
}

func (v *AV1) Type() codecs.CodecType {
	return Type{}
}

func (v *AV1) Equals(codec codecs.Codec) bool {
	if codec == nil {
		return false
	}
	AV1Codec, ok := codec.(*AV1)
	if !ok {
		return false
	}
	if v.CodecType() != AV1Codec.CodecType() || v.MediaType() != AV1Codec.MediaType() {
		return false
	}
	if v.Width() != AV1Codec.Width() || v.Height() != AV1Codec.Height() {
		return false
	}
	return true
}

func (v *AV1) String() string {
	return fmt.Sprintf("AV1. Width: %d, Height: %d", v.width, v.height)
}

func (v *AV1) MimeType() string {
	return pion.MimeTypeAV1
}

func (v *AV1) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (v *AV1) CodecType() types.CodecType {
	return types.CodecTypeAV1
}

func (v *AV1) Width() int {
	return v.width
}

func (v *AV1) Height() int {
	return v.height
}

func (v *AV1) ClockRate() uint32 {
	return 90000
}

func (v *AV1) FPS() float64 {
	return 30
}

func (v *AV1) PixelFormat() int {
	return avutil.AV_PIX_FMT_YUV420P
}

// ExtraData use readonly
func (v *AV1) ExtraData() []byte {
	return nil
}

func (v *AV1) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(v.Type().AVCodecID())
	codecCtx.SetCodecType(v.Type().AVMediaType())
	codecCtx.SetWidth(v.Width())
	codecCtx.SetHeight(v.Height())
	codecCtx.SetTimeBase(avutil.NewRational(1, int(v.FPS())))
	codecCtx.SetPixelFormat(avutil.PixelFormat(v.PixelFormat()))
	codecCtx.SetExtraData(v.ExtraData())
}

func (v *AV1) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     v.MimeType(),
		ClockRate:    v.ClockRate(),
		Channels:     0,
		SDPFmtpLine:  "level-idx=5;profile=0;tier=0",
		RTCPFeedback: nil,
	}, nil
}

func (v *AV1) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	payloadType := 97
	return engines.RTPCodecParameters{
		PayloadType: uint8(payloadType),
		CodecType:   v.CodecType(),
		ClockRate:   90000,
		MediaDescription: sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media: v.MediaType().String(),
				Port: sdp.RangedPort{
					Value: targetPort,
				},
				Protos:  []string{"RTP", "AVP"},
				Formats: []string{fmt.Sprintf("%d", payloadType)},
			},
			Attributes: []sdp.Attribute{
				{
					Key:   "rtpmap",
					Value: fmt.Sprintf("%d %s/%d", payloadType, strings.ToLower(string(v.CodecType())), v.ClockRate()),
				},
			},
		},
	}, nil
}

func (v *AV1) SetVideoTranscodeInfo(info codecs.VideoTranscodeInfo) {
	return
}

func (v *AV1) GetBitStreamFilter() bitstreamfilter.BitStreamFilter {
	return bitstreamfilter.NewBitStream(v.CodecType())
}
