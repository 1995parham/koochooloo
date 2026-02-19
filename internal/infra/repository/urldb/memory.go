package urldb

import (
	"context"
	"iter"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
)

type MemoryURL struct {
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
		for k, v := range m.store {
			if v.ExpireTime == nil || v.ExpireTime.After(time.Now()) {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

func (m *MemoryURL) Inc(_ context.Context, key string) error {
	u, ok := m.live(key)
	if !ok {
		return urlrepo.ErrKeyNotFound
	}

	u.Count++
	m.store[key] = u

	return nil
}

func (m *MemoryURL) Set(_ context.Context, key string, url string, expire *time.Time, count int) error {
	if _, ok := m.store[key]; ok {
		return urlrepo.ErrDuplicateKey
	}

	m.store[key] = model.URL{
		Key:        key,
		URL:        url,
		Count:      count,
		ExpireTime: expire,
	}

	return nil
}

func (m *MemoryURL) Get(_ context.Context, key string) (string, error) {
	if u, ok := m.live(key); ok {
		return u.URL, nil
	}

	return "", urlrepo.ErrKeyNotFound
}

func (m *MemoryURL) Count(_ context.Context, key string) (int, error) {
	if u, ok := m.live(key); ok {
		return u.Count, nil
	}

	return 0, urlrepo.ErrKeyNotFound
}

func (m *MemoryURL) live(key string) (model.URL, bool) {
	u, ok := m.store[key]
	if !ok {
		return u, false
	}

	if u.ExpireTime != nil && u.ExpireTime.Before(time.Now()) {
		return u, false
	}

	return u, true
}
