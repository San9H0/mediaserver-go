package peertopeer

import (
	"context"
	"sync"
)

type Server struct {
	mu sync.RWMutex

	sessions map[string]*Session
}

func NewServer() *Server {
	return &Server{
		sessions: make(map[string]*Session),
	}
}

func (s *Server) AddSession(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
}

func (s *Server) GetSession(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessions[id]
}

func (s *Server) RemoveSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sess.ID)
}

func (s *Server) GetSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *Server) Run(ctx context.Context) {
	for {
		sess := s.GetSessions()
		if len(sess) > 0 {

		}
	}
}
