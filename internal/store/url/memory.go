package url

import (
	"context"
	"time"

	"github.com/1995parham/koochooloo/internal/model"
)

type MemoryURL struct {
	store map[string]model.URL
}

func NewMemoryURL() *MemoryURL {
	return &MemoryURL{
		store: make(map[string]model.URL),
	}
}

func (m MemoryURL) Inc(ctx context.Context, key string) error {
	u, ok := m.store[key]
	if !ok {
		return ErrKeyNotFound
	}

	u.Count++
	m.store[key] = u

	return nil
}

func (m MemoryURL) Set(ctx context.Context, key string, url string, expire *time.Time, count int) (string, error) {
	if key == "" {
		key = model.Key()
	} else {
		key = "$" + key
	}

	if _, ok := m.store[key]; ok {
		return "", ErrDuplicateKey
	}

	m.store[key] = model.URL{
		Key:        key,
		URL:        url,
		Count:      count,
		ExpireTime: expire,
	}

	return key, nil
}

func (m MemoryURL) Get(ctx context.Context, key string) (string, error) {
	url := m.store[key]

	if url.ExpireTime == nil || url.ExpireTime.After(time.Now()) {
		url.Count++

		return url.URL, nil
	}

	return "", ErrKeyNotFound
}

func (m MemoryURL) Count(ctx context.Context, key string) (int, error) {
	url, found := m.store[key]
	if !found {
		return 0, ErrKeyNotFound
	}

	return url.Count, nil
}
