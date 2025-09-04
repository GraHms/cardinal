package store

import (
	"context"
	"sync"
	"time"

	"github.com/grahms/cardinal/core"
)

type InMemory struct {
	mu     sync.Mutex
	data   map[string]item
	defTTL time.Duration
}

type item struct {
	val map[string]any
	exp time.Time
}

func NewInMemoryStore(defaultTTL time.Duration) *InMemory {
	m := &InMemory{data: make(map[string]item), defTTL: defaultTTL}
	go m.gc()
	return m
}

func (m *InMemory) Get(_ context.Context, sid string) (map[string]any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.data[sid]
	if !ok || time.Now().After(it.exp) {
		return map[string]any{}, nil
	}
	return clone(it.val), nil
}

func (m *InMemory) Put(_ context.Context, sid string, d map[string]any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ttl <= 0 {
		ttl = m.defTTL
	}
	m.data[sid] = item{val: clone(d), exp: time.Now().Add(ttl)}
	return nil
}

func (m *InMemory) Del(_ context.Context, sid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, sid)
	return nil
}

func (m *InMemory) gc() {
	t := time.NewTicker(time.Minute)
	for range t.C {
		now := time.Now()
		m.mu.Lock()
		for k, it := range m.data {
			if now.After(it.exp) {
				delete(m.data, k)
			}
		}
		m.mu.Unlock()
	}
}

func clone(m0 map[string]any) map[string]any {
	if m0 == nil {
		return map[string]any{}
	}
	cp := make(map[string]any, len(m0))
	for k, v := range m0 {
		cp[k] = v
	}
	return cp
}

// Ensure interface conformance at compile-time (when built).
var _ core.Store = (*InMemory)(nil)
