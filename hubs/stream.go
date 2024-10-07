package hubs

import (
	"mediaserver-go/codecs"
	"mediaserver-go/utils/types"
	"sync"
)

type Stream struct {
	mu sync.RWMutex

	subscribers []chan *HubSource

	source []*HubSource
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) GetCodecs() map[types.MediaType]codecs.Codec {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[types.MediaType]codecs.Codec)

	for _, source := range s.source {
		codec, err := source.Codec()
		if err != nil {
			continue
		}
		result[source.MediaType()] = codec
	}
	return result
}

func (s *Stream) Subscribe() chan *HubSource {
	ch := make(chan *HubSource, 10)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscribers = append(s.subscribers, ch)
	for _, source := range s.source {
		ch <- source
	}
	return ch
}

func (s *Stream) AddSource(source *HubSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source = append(s.source, source)

	for _, subscriber := range s.subscribers {
		subscriber <- source
	}
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
