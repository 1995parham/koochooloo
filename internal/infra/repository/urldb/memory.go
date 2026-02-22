package urldb

import (
	"context"
	"iter"
	"sync"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
)

type MemoryURL struct {
	mu    sync.RWMutex
	store map[string]model.URL
}

func ProvideMemory() *MemoryURL {
	return &MemoryURL{
		store: make(map[string]model.URL),
	}
}

// All returns an iterator over all non-expired URLs in the store.
func (m *MemoryURL) All() iter.Seq2[string, model.URL] {
	return func(yield func(string, model.URL) bool) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		for k, v := range m.store {
			if !v.IsExpired() {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

func (m *MemoryURL) IncrementCount(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.liveLocked(key)
	if !ok {
		return urlrepo.ErrKeyNotFound
	}

	u.Count++
	m.store[key] = u

	return nil
}

func (m *MemoryURL) Save(_ context.Context, url model.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.store[url.Key]; ok {
		return urlrepo.ErrDuplicateKey
	}

	m.store[url.Key] = url

	return nil
}

func (m *MemoryURL) FindByKey(_ context.Context, key string) (model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if u, ok := m.liveLocked(key); ok {
		return u, nil
	}

	return model.URL{}, urlrepo.ErrKeyNotFound
}

// liveLocked checks if a key exists and is not expired. Caller must hold mu.
func (m *MemoryURL) liveLocked(key string) (model.URL, bool) {
	u, ok := m.store[key]
	if !ok {
		return u, false
	}

	if u.IsExpired() {
		return u, false
	}

	return u, true
}
