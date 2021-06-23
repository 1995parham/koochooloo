package url

import (
	"context"
	"time"
)

// URL stores and retrieves urls.
type URL interface {
	Inc(ctx context.Context, key string) error
	Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Count(ctx context.Context, key string) (int, error)
}
