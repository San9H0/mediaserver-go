package images

import (
	"context"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/units"
	"sync"
)

type Handler struct {
	mu sync.RWMutex

	encoding   string
	negotiated []hubs.Track
}

func NewHandler(encoding string) *Handler {
	return &Handler{
		encoding: encoding,
	}
}

func (h *Handler) NegotiatedTracks() []hubs.Track {
	ret := make([]hubs.Track, 0, len(h.negotiated))
	return append(ret, h.negotiated...)
}

func (h *Handler) Init(ctx context.Context, sources []*hubs.HubSource) error {
	var negotiated []hubs.Track
	for _, source := range sources {
		codec, err := source.Codec()
		if err != nil {
			return err
		}
		negotiated = append(negotiated, source.GetTrack(codec))
	}
	h.negotiated = negotiated
	return nil
}

func (h *Handler) OnClosed(ctx context.Context) error {
	return nil
}

func (h *Handler) OnTrack(ctx context.Context, track hubs.Track) (*TrackContext, error) {
	return nil, nil
}

func (h *Handler) OnVideo(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {

	return nil
}

func (h *Handler) OnAudio(ctx context.Context, trackCtx *TrackContext, unit units.Unit) error {
	return nil
}
