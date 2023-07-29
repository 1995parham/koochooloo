package urldb

import (
	"context"
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

func (m *MemoryURL) Inc(_ context.Context, key string) error {
	u, ok := m.store[key]
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
	url, ok := m.store[key]
	if ok && (url.ExpireTime == nil || url.ExpireTime.After(time.Now())) {
		return url.URL, nil
	}

	return "", urlrepo.ErrKeyNotFound
}

func (m *MemoryURL) Count(_ context.Context, key string) (int, error) {
	url, ok := m.store[key]
	if ok && (url.ExpireTime == nil || url.ExpireTime.After(time.Now())) {
		return url.Count, nil
	}

	return 0, urlrepo.ErrKeyNotFound
}
