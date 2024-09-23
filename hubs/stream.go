package hubs

import (
	"mediaserver-go/utils/types"
	"sync"
)

type Stream struct {
	mu sync.RWMutex

	source []*HubSource
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) AddSource(source *HubSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source = append(s.source, source)
}

func (s *Stream) RemoveSource(source *HubSource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.source {
		if t == source {
			s.source = append(s.source[:i], s.source[i+1:]...)
			return
		}
	}
}

func (s *Stream) Sources() []*HubSource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tracks := make([]*HubSource, 0, len(s.source))
	return append(tracks, s.source...)
}

func (s *Stream) SourcesMap() map[types.MediaType]*HubSource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sources := make(map[types.MediaType]*HubSource)
	for _, t := range s.source {
		sources[t.MediaType()] = t
	}
	return sources
}

func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.source {
		t.Close()
	}
	s.source = nil
}
