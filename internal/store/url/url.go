package url

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrKeyNotFound indicates that given key does not exist on database.
	ErrKeyNotFound = errors.New("given key does not exist or expired")
	// ErrDuplicateKey indicates that given key is exists on database.
	ErrDuplicateKey = errors.New("given key is exist")
)

// URL stores and retrieves urls.
type URL interface {
	Inc(ctx context.Context, key string) error
	Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Count(ctx context.Context, key string) (int, error)
}
