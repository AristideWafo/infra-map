package cache

import (
	"sync"
	"time"
)

// Memory est un cache TTL en mémoire — l'unique stockage autorisé (ADR-002).
type Memory struct {
	mu    sync.RWMutex
	items map[string]*item
	ttl   time.Duration
	now   func() time.Time
}

type item struct {
	value     interface{}
	storedAt  time.Time
	expiresAt time.Time
}

func New(ttl time.Duration) *Memory {
	return &Memory{items: make(map[string]*item), ttl: ttl, now: time.Now}
}

func (m *Memory) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.now()
	m.items[key] = &item{value: value, storedAt: n, expiresAt: n.Add(m.ttl)}
}

func (m *Memory) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.items[key]
	if !ok || m.now().After(it.expiresAt) {
		return nil, false
	}
	return it.value, true
}

// Age retourne l'âge en secondes de l'entrée, ou -1 si absente.
// Une entrée expirée a quand même un âge — utile pour signaler des données périmées.
func (m *Memory) Age(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.items[key]
	if !ok {
		return -1
	}
	return int(m.now().Sub(it.storedAt).Seconds())
}
