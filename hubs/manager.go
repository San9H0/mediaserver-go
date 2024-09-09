package hubs

import (
	"mediaserver-go/utils/types"
	"sync"
)

type Manager struct {
	mu sync.RWMutex

	hubs map[string]Hub
}

func NewManager() *Manager {
	return &Manager{
		hubs: make(map[string]Hub),
	}
}

func (m *Manager) NewHub(path string, mediaType types.MediaType) (Hub, error) {
	h, err := NewHub(mediaType)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if h, ok := m.hubs[path]; ok {
		return h, nil
	}
	m.hubs[path] = h
	return h, nil
}

func (m *Manager) AddHub(name string, hub Hub) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hubs[name] = hub
}

func (m *Manager) GetHub(name string) (Hub, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hub, ok := m.hubs[name]
	return hub, ok
}

func (m *Manager) RemoveHub(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.hubs, name)
}

func (m *Manager) Hubs() map[string]Hub {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hubs := make(map[string]Hub)
	for k, v := range m.hubs {
		hubs[k] = v
	}
	return hubs
}
