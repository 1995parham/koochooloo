package urlrepo

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

// Repository stores and retrieves urls.
type Repository interface {
	Inc(context.Context, string) error
	Set(context.Context, string, string, *time.Time, int) (string, error)
	Get(context.Context, string) (string, error)
	Count(context.Context, string) (int, error)
}
