package hubs

import (
	"mediaserver-go/utils/types"
	"sync"
)

type Stream struct {
	mu sync.RWMutex

	tracks []*Track
}

func NewStream() *Stream {
	return &Stream{}
}

func (s *Stream) AddTrack(track *Track) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks = append(s.tracks, track)
}

func (s *Stream) RemoveTrack(track *Track) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.tracks {
		if t == track {
			s.tracks = append(s.tracks[:i], s.tracks[i+1:]...)
			return
		}
	}
}

func (s *Stream) Tracks() []*Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tracks := make([]*Track, 0, len(s.tracks))
	return append(tracks, s.tracks...)
}

func (s *Stream) TracksMap() map[types.MediaType]*Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tracks := make(map[types.MediaType]*Track)
	for _, t := range s.tracks {
		tracks[t.MediaType()] = t
	}
	return tracks
}

func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.tracks {
		t.Close()
	}
	s.tracks = nil
}
