package hubs

import (
	"errors"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
	"sync/atomic"
)

var (
	invalidCodecType = errors.New("invalid codec type")
)

type Hub interface {
	PushUnit(unit units.Unit)
	Register(chan units.Unit)
}

func NewHub(mediaType types.MediaType) (Hub, error) {
	switch mediaType {
	case types.MediaTypeVideo:
		return &hub{}, nil
	case types.MediaTypeAudio:
		return &hub{}, nil
	default:
		return nil, invalidCodecType
	}
}

type hub struct {
	mu sync.RWMutex

	ready atomic.Bool
	param atomic.Pointer[Parameter]

	channels []chan units.Unit
}

type Parameter struct {
	MediaType  types.MediaType
	CodecType  types.CodecType
	Width      int
	Height     int
	Bitrate    int
	SampleRate int
}

func (h *hub) SetCodecParameter(param Parameter) {
	h.param.Store(&param)
	h.ready.Store(true)
}

func (h *hub) PushUnit(u units.Unit) {
	if !h.ready.Load() {
		return
	}
	h.mu.RLock()
	chs := h.channels
	h.mu.RUnlock()
	for _, ch := range chs {
		utils.SendOrDrop(ch, u)
	}
}

func (h *hub) Register(ch chan units.Unit) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.channels = append(h.channels, ch)
}
