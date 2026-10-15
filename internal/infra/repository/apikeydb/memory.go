package apikeydb

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/apikeyrepo"
)

// MemoryAPIKey is an in-memory implementation of apikeyrepo.Repository.
type MemoryAPIKey struct {
	mu     sync.RWMutex
	nextID uint
	store  map[uint]model.APIKey
}

// ProvideMemory creates an empty in-memory API key store.
func ProvideMemory() *MemoryAPIKey {
	return &MemoryAPIKey{
		mu:     sync.RWMutex{},
		nextID: 0,
		store:  make(map[uint]model.APIKey),
	}
}

// Create persists a new key and returns it with its assigned ID.
func (m *MemoryAPIKey) Create(_ context.Context, key model.APIKey) (model.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	key.ID = m.nextID
	m.store[key.ID] = key

	return key, nil
}

// FindByHash returns the key whose SHA-256 hash is hash.
func (m *MemoryAPIKey) FindByHash(_ context.Context, hash string) (model.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, k := range m.store {
		if k.Hash == hash {
			return k, nil
		}
	}

	return model.APIKey{}, apikeyrepo.ErrAPIKeyNotFound
}

// List returns all keys, ordered by id.
func (m *MemoryAPIKey) List(_ context.Context) ([]model.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]model.APIKey, 0, len(m.store))
	for _, k := range m.store {
		keys = append(keys, k)
	}

	slices.SortFunc(keys, func(a, b model.APIKey) int { return int(a.ID) - int(b.ID) })

	return keys, nil
}

// Revoke marks the key with the given id revoked at the given time.
func (m *MemoryAPIKey) Revoke(_ context.Context, id uint, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	k, ok := m.store[id]
	if !ok {
		return apikeyrepo.ErrAPIKeyNotFound
	}

	k.RevokedAt = &at
	m.store[id] = k

	return nil
}

// Touch records that the key with the given id was used at the given time.
func (m *MemoryAPIKey) Touch(_ context.Context, id uint, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if k, ok := m.store[id]; ok {
		k.LastUsedAt = &at
		m.store[id] = k
	}

	return nil
}
