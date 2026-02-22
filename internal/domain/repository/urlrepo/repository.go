package urlrepo

import (
	"context"
	"errors"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

var (
	// ErrKeyNotFound indicates that given key does not exist on database.
	ErrKeyNotFound = errors.New("given key does not exist or expired")
	// ErrDuplicateKey indicates that given key is exists on database.
	ErrDuplicateKey = errors.New("given key is exist")
)

// Repository stores and retrieves urls.
type Repository interface {
	Save(ctx context.Context, url model.URL) error
	FindByKey(ctx context.Context, key string) (model.URL, error)
	IncrementCount(ctx context.Context, key string) error
}
