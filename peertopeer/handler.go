package peertopeer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

type ServiceHandler struct {
	mu sync.RWMutex

	handlerCh  chan []byte
	remotePair map[*Session]*Session
}

func NewMessageHandler() *ServiceHandler {
	return &ServiceHandler{
		handlerCh:  make(chan []byte),
		remotePair: make(map[*Session]*Session),
	}
}

func (m *ServiceHandler) AddPair(sess1, sess2 *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.remotePair[sess1] = sess2
	m.remotePair[sess2] = sess1
}

func (m *ServiceHandler) RemovePair(sess *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if remote, ok := m.remotePair[sess]; ok {
		delete(m.remotePair, sess)
		delete(m.remotePair, remote)
	}
}

func (m *ServiceHandler) HandleMessage(sess *Session, data []byte) error {
	header := &Header{}
	if err := json.Unmarshal(data, header); err != nil {
		return err
	}

	switch header.Type {
	case "candidate":
		remote, ok := m.remotePair[sess]
		if !ok {
			return errors.New("no remote pair")
		}

		candidate := string(data)
		candidate = strings.ReplaceAll(candidate, "192.168.219.104", "192.168.0.123")
		candidate = strings.ReplaceAll(candidate, "172.17.128.1", "172.17.128.123")
		return remote.WriteData(context.Background(), []byte(candidate))
	case "offer", "answer":
		remote, ok := m.remotePair[sess]
		if !ok {
			return errors.New("no remote pair")
		}
		return remote.WriteData(context.Background(), data)
	}
	return nil
}
