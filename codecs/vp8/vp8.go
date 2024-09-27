package vp8

import (
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
	"strings"
)

type VP8 struct {
	Base

	config *Config
}

func NewVP8(config *Config) *VP8 {
	return &VP8{
		Base:   Base{},
		config: config,
	}
}

func (v *VP8) Equals(codec codecs.Codec) bool {
	if codec == nil {
		return false
	}
	vp8Codec, ok := codec.(*VP8)
	if !ok {
		return false
	}
	if v.Width() != vp8Codec.Width() || v.Height() != vp8Codec.Height() {
		return false
	}
	return true
}
func (v *VP8) String() string {
	return v.MimeType()
}

func (v *VP8) GetBase() codecs.Base {
	return v.Base
}

func (v *VP8) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (v *VP8) CodecType() types.CodecType {
	return types.CodecTypeVP8
}

func (v *VP8) Width() int {
	return v.config.Width
}

func (v *VP8) Height() int {
	return v.config.Height
}

func (v *VP8) ClockRate() uint32 {
	return 90000
}

func (v *VP8) FPS() float64 {
	return 30
}

func (v *VP8) PixelFormat() int {
	return avutil.AV_PIX_FMT_YUV420P
}

// ExtraData use readonly
func (v *VP8) ExtraData() []byte {
	return nil
}

func (v *VP8) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(v.AVCodecID())
	codecCtx.SetCodecType(v.AVMediaType())
	codecCtx.SetWidth(v.Width())
	codecCtx.SetHeight(v.Height())
	codecCtx.SetTimeBase(avutil.NewRational(1, int(v.FPS())))
	codecCtx.SetPixelFormat(avutil.PixelFormat(v.PixelFormat()))
	codecCtx.SetExtraData(v.ExtraData())
}

func (v *VP8) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     v.MimeType(),
		ClockRate:    v.ClockRate(),
		Channels:     0,
		SDPFmtpLine:  "",
		RTCPFeedback: nil,
	}, nil
}

func (v *VP8) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
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

func (v *VP8) SetVideoTranscodeInfo(info codecs.VideoTranscodeInfo) {
	return
}

func (v *VP8) GetBitStreamFilter() bitstreamfilter.BitStreamFilter {
	return bitstreamfilter.NewBitStream(v.CodecType())
}
