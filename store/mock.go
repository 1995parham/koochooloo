package store

import (
	"context"
	"fmt"
	"time"

	"github.com/1995parham/koochooloo/model"
)

type MockURL struct {
	store map[string]model.URL
}

func NewMockURL() *MockURL {
	return &MockURL{
		store: make(map[string]model.URL),
	}
}

func (m MockURL) Inc(ctx context.Context, key string) error {
	return nil
}

func (m MockURL) Set(ctx context.Context, key string, url string, expire *time.Time, count int) (string, error) {
	if key == "" {
		key = Key()
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

	fmt.Println(count)

	return key, nil
}

func (m MockURL) Get(ctx context.Context, key string) (string, error) {
	url := m.store[key]

	if url.ExpireTime == nil || url.ExpireTime.After(time.Now()) {
		return url.URL, nil
	}

	return "", ErrKeyNotFound
}

func (m MockURL) Count(ctx context.Context, key string) (int, error) {
	url := m.store[key]

	//if url.ExpireTime == nil || url.ExpireTime.After(time.Now()) {
	//	fmt.Println(url.Count)
	//
	//}

	return url.Count, nil
}
