package hubs

import (
	"go.uber.org/zap"
	"mediaserver-go/utils/log"
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	streams map[ /*streamID*/ string]*Stream
}

func NewHub() *Hub {
	return &Hub{
		streams: make(map[string]*Stream),
	}
}

func (h *Hub) AddStream(id string, stream *Stream) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.streams[id] = stream
	log.Logger.Info("stream added", zap.String("streamkey", id))
}

func (h *Hub) RemoveStream(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.streams, id)
	log.Logger.Info("stream deleted", zap.String("streamkey", id))
}

func (h *Hub) GetStream(id string) (*Stream, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stream, ok := h.streams[id]
	return stream, ok
}
