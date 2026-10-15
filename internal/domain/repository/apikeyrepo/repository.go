// Package apikeyrepo defines the persistence contract for API keys.
package apikeyrepo

import (
	"context"
	"errors"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

// ErrAPIKeyNotFound indicates that no API key matches the given lookup.
var ErrAPIKeyNotFound = errors.New("api key does not exist")

// Repository stores and retrieves API keys.
type Repository interface {
	// Create persists a new key and returns it with its assigned ID.
	Create(ctx context.Context, key model.APIKey) (model.APIKey, error)
	// FindByHash returns the key whose SHA-256 hash is hash.
	FindByHash(ctx context.Context, hash string) (model.APIKey, error)
	// List returns all keys, revoked ones included.
	List(ctx context.Context) ([]model.APIKey, error)
	// Revoke marks the key with the given id revoked at the given time.
	Revoke(ctx context.Context, id uint, at time.Time) error
	// Touch records that the key with the given id was used at the given time.
	Touch(ctx context.Context, id uint, at time.Time) error
}
