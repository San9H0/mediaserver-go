package codecs

import (
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils/types"
	"strings"
)

var _ VideoCodec = (*VP8)(nil)

type VP8 struct {
	width, height int
}

func NewVP8(width, height int) (*VP8, error) {
	return &VP8{
		width: width, height: height,
	}, nil
}

func (v *VP8) Equals(codec Codec) bool {
	if codec == nil {
		return false
	}
	vp8Codec, ok := codec.(*VP8)
	if !ok {
		return false
	}
	if v.CodecType() != vp8Codec.CodecType() || v.MediaType() != vp8Codec.MediaType() {
		return false
	}
	if v.Width() != vp8Codec.Width() || v.Height() != vp8Codec.Height() {
		return false
	}
	return true
}

func (v *VP8) String() string {
	return fmt.Sprintf("VP8. Width: %d, Height: %d", v.width, v.height)
}

func (v *VP8) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (v *VP8) CodecType() types.CodecType {
	return types.CodecTypeVP8
}

func (v *VP8) Width() int {
	return v.width
}

func (v *VP8) Height() int {
	return v.height
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
	codecCtx.SetCodecID(types.CodecIDFromType(v.CodecType()))
	codecCtx.SetCodecType(types.MediaTypeToFFMPEG(v.MediaType()))
	codecCtx.SetWidth(v.Width())
	codecCtx.SetHeight(v.Height())
	codecCtx.SetTimeBase(avutil.NewRational(1, int(v.FPS())))
	codecCtx.SetPixelFormat(avutil.PixelFormat(v.PixelFormat()))
	codecCtx.SetExtraData(v.ExtraData())
}

func (v *VP8) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     types.MimeTypeFromCodecType(v.CodecType()),
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
