package codecs

import (
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/parser/format"
	"mediaserver-go/utils/types"
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
	fps           float64
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
		fps:       spsSet.FPS(),
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

func (h *H264) Height() int {
	return h.height
}

func (h *H264) FPS() float64 {
	return h.fps
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

func (h *H264) Profile() string {
	profileIdc := h.extraData[0]
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
