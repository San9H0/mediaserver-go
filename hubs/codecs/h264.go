package codecs

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/engines"
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

	config *H264Config
}

func NewH264FromConfig(config *H264Config) (Codec, error) {
	if len(config.SPS) == 0 || len(config.PPS) == 0 {
		return nil, errInvalidSPSLength
	}
	pps := config.PPS
	sps := config.SPS
	spsSet := &h264.SPS{}
	if err := spsSet.Unmarshal(sps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal. %w", err)
	}

	return &H264{
		sps:      append(make([]byte, 0, len(sps)), sps...),
		pps:      append(make([]byte, 0, len(pps)), pps...),
		spsSet:   spsSet,
		width:    spsSet.Width(),
		height:   spsSet.Height(),
		pixelFmt: makePixelFmt(spsSet),
		config:   config,
	}, nil
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

	config := &H264Config{}
	if err := config.UnmarshalFromSPSPPS(sps, pps); err != nil {
		return nil, err
	}
	return &H264{
		sps:      append(make([]byte, 0, len(sps)), sps...),
		pps:      append(make([]byte, 0, len(pps)), pps...),
		spsSet:   spsSet,
		width:    spsSet.Width(),
		height:   spsSet.Height(),
		pixelFmt: makePixelFmt(spsSet),
		config:   config,
	}, nil
}

func (h *H264) Equals(codec Codec) bool {
	if codec == nil {
		return false
	}
	h264Codec, ok := codec.(*H264)
	if !ok {
		return false
	}
	if h.CodecType() != h264Codec.CodecType() || h.MediaType() != h264Codec.MediaType() {
		return false
	}
	if h.Width() != h264Codec.Width() || h.Height() != h264Codec.Height() || h.PixelFormat() != h264Codec.PixelFormat() {
		return false
	}
	if !bytes.Equal(h.SPS(), h264Codec.SPS()) || bytes.Equal(h.PPS(), h264Codec.PPS()) {
		return false
	}
	return true
}

func (h *H264) String() string {
	return fmt.Sprintf("H264. Width: %d, Height: %d, PixelFmt: %d", h.width, h.height, h.pixelFmt)
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
	b, _ := h.config.Marshal()
	return b
}

// SPS use readonly
func (h *H264) SPS() []byte {
	return h.sps
}

// PPS use readonly
func (h *H264) PPS() []byte {
	return h.pps
}

func (h *H264) profileIDC() uint8 {
	return uint8(h.config.ProfileID)
}

func (h *H264) constraintFlag() uint8 {
	return uint8(h.config.ProfileComp)
}

func (h *H264) level() uint8 {
	return uint8(h.config.LevelID)
}

func (h *H264) profile() string {
	return fmt.Sprintf("%02x%02x%02x", h.config.ProfileID, h.config.ProfileComp, h.config.LevelID)
}

func makePixelFmt(spsSet *h264.SPS) int {
	fmt.Println("[TESTDEBUG] makePixelFmt.. spsSet.ChromaFormatIdc:", spsSet.ChromaFormatIdc)
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
	codecCtx.SetProfile(int(h.profileIDC()))
	codecCtx.SetLevel(int(h.level()))
	codecCtx.SetExtraData(h.ExtraData())

	fmt.Println("[TESTDEBUG] SetCodecID:", types.CodecIDFromType(h.CodecType()))
	fmt.Println("[TESTDEBUG] SetCodecType:", types.MediaTypeToFFMPEG(h.MediaType()))
	fmt.Println("[TESTDEBUG] SetWidth:", h.Width())
	fmt.Println("[TESTDEBUG] SetHeight:", h.Height())
	fmt.Println("[TESTDEBUG] SetTimeBase:", avutil.NewRational(1, int(h.FPS())))
	fmt.Println("[TESTDEBUG] SetPixelFormat:", h.PixelFormat())
	fmt.Println("[TESTDEBUG] SetProfile:", int(h.profileIDC()))
	fmt.Println("[TESTDEBUG] SetLevel:", int(h.level()))
	fmt.Println("[TESTDEBUG] SetExtraData:", h.ExtraData())
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
