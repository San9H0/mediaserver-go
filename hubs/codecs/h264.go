package codecs

import (
	"mediaserver-go/parser/format"
	"mediaserver-go/utils/types"
	"sync"
)

var _ VideoCodec = (*H264)(nil)

type H264 struct {
	mu sync.RWMutex

	meta H264Metadata
}

func NewH264() *H264 {
	return &H264{}
}

type H264Metadata struct {
	CodecType types.CodecType //codecCtx.SetCodecID(inputCodecpar.CodecID())
	MediaType types.MediaType //codecCtx.SetCodecType(inputCodecpar.CodecType())
	Width     int             //codecCtx.SetWidth(inputCodecpar.Width())
	Height    int             //codecCtx.SetHeight(inputCodecpar.Height())
	FPS       float64         //codecCtx.SetTimeBase(avutil.NewRational(1, 30))
	PixelFmt  int             //codecCtx.SetPixelFormat(avutil.AV_PIX_FMT_YUV420P)
	SPS       []byte          //extradata := format.ExtraDataForAVCC(sps, pps), //codecCtx.SetExtraData(extradata)
	PPS       []byte
}

func (h *H264) SetMetaData(meta H264Metadata) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.meta = meta
}

func (h *H264) CodecType() types.CodecType {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.CodecType
}

func (h *H264) MediaType() types.MediaType {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.MediaType
}

func (h *H264) Width() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.Width
}

func (h *H264) Height() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.Height
}

func (h *H264) FPS() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.FPS
}

func (h *H264) PixelFormat() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.meta.PixelFmt
}

func (h *H264) ExtraData() []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return format.ExtraDataForAVCC(h.meta.SPS, h.meta.PPS)
}
