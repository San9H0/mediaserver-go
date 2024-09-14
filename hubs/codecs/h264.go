package codecs

import (
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/parser/format"
	"mediaserver-go/utils/types"
	"strings"
)

var (
	errInvalidSPSLength     = errors.New("invalid SPS length")
	errInvalidPPSLength     = errors.New("invalid PPS length")
	errFailedToUnmarshalSPS = errors.New("failed to unmarshal SPS")
)

var _ VideoCodec = (*H264)(nil)

type H264 struct {
	sps, pps      []byte
	spsSet        *h264.SPS
	width, height int
	pixelFmt      int
	extraData     []byte
}

func NewH264(sps, pps []byte) (*H264, error) {
	if len(sps) == 0 {
		return nil, errInvalidSPSLength
	}
	if len(pps) == 0 {
		return nil, errInvalidPPSLength
	}

	spsSet := &h264.SPS{}
	if err := spsSet.Unmarshal(sps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal. %w", err)
	}

	return &H264{
		sps:       append(make([]byte, 0, len(sps)), sps...),
		pps:       append(make([]byte, 0, len(pps)), pps...),
		spsSet:    spsSet,
		width:     spsSet.Width(),
		height:    spsSet.Height(),
		pixelFmt:  makePixelFmt(spsSet),
		extraData: format.ExtraDataForAVC(sps, pps),
	}, nil
}

func (h *H264) MediaType() types.MediaType {
	return types.MediaTypeVideo
}

func (h *H264) CodecType() types.CodecType {
	return types.CodecTypeH264
}

func (h *H264) Width() int {
	return h.width
}

func (h *H264) ClockRate() uint32 {
	return 90000
}

func (h *H264) Height() int {
	return h.height
}

func (h *H264) FPS() float64 {
	return 30
}

func (h *H264) PixelFormat() int {
	return h.pixelFmt
}

// ExtraData use readonly
func (h *H264) ExtraData() []byte {
	return h.extraData
}

// SPS use readonly
func (h *H264) SPS() []byte {
	return h.sps
}

// PPS use readonly
func (h *H264) PPS() []byte {
	return h.pps
}

func (h *H264) profile() string {
	profileIdc := h.extraData[1]
	profileCompatibility := h.extraData[2]
	levelIdc := h.extraData[3]
	return fmt.Sprintf("%02x%02x%02x", profileIdc, profileCompatibility, levelIdc)
}

func makePixelFmt(spsSet *h264.SPS) int {
	switch spsSet.ChromaFormatIdc {
	case 0:
		return avutil.AV_PIX_FMT_GRAY8
	case 1:
		return avutil.AV_PIX_FMT_YUV420P
	case 2:
		return avutil.AV_PIX_FMT_YUV422P
	case 3:
		return avutil.AV_PIX_FMT_YUV444P
	default:
		return avutil.AV_PIX_FMT_NONE
	}
}

func (h *H264) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(types.CodecIDFromType(h.CodecType()))
	codecCtx.SetCodecType(types.MediaTypeToFFMPEG(h.MediaType()))
	codecCtx.SetWidth(h.Width())
	codecCtx.SetHeight(h.Height())
	codecCtx.SetTimeBase(avutil.NewRational(1, int(h.FPS())))
	codecCtx.SetPixelFormat(avutil.PixelFormat(h.PixelFormat()))
	codecCtx.SetExtraData(h.ExtraData())
}

func (h *H264) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     types.MimeTypeFromCodecType(h.CodecType()),
		ClockRate:    h.ClockRate(),
		Channels:     0,
		SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=%s", h.profile()),
		RTCPFeedback: nil,
	}, nil
}

func (h *H264) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	payloadType := 98
	return engines.RTPCodecParameters{
		PayloadType: uint8(payloadType),
		ClockRate:   90000,
		CodecType:   h.CodecType(),
		MediaDescription: sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media: h.MediaType().String(),
				Port: sdp.RangedPort{
					Value: targetPort,
				},
				Protos:  []string{"RTP", "AVP"},
				Formats: []string{fmt.Sprintf("%d", payloadType)},
			},
			Attributes: []sdp.Attribute{
				{
					Key:   "rtpmap",
					Value: fmt.Sprintf("%d %s/%d", payloadType, strings.ToUpper(string(h.CodecType())), h.ClockRate()),
				},
				{
					Key:   "fmtp",
					Value: fmt.Sprintf("%d %s", payloadType, fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=%s", h.profile())),
				},
			},
		},
	}, nil
}
