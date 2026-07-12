// Package userrepo defines the persistence contract for user accounts.
package userrepo

import (
	"context"
	"errors"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

var (
	// ErrUserNotFound indicates that no user matches the given lookup.
	ErrUserNotFound = errors.New("user does not exist")
	// ErrDuplicateUsername indicates that the username is already taken.
	ErrDuplicateUsername = errors.New("username already exists")
)

// Repository stores and retrieves user accounts.
type Repository interface {
	// Create persists a new user and returns it with its assigned ID.
	Create(ctx context.Context, user model.User) (model.User, error)
	// FindByUsername returns the user with the given username.
	FindByUsername(ctx context.Context, username string) (model.User, error)
	// FindByID returns the user with the given id.
	FindByID(ctx context.Context, id uint) (model.User, error)
	// FindBySubject returns the user federated from provider with the given subject.
	FindBySubject(ctx context.Context, provider model.Provider, subject string) (model.User, error)
	// List returns all users.
	List(ctx context.Context) ([]model.User, error)
	// SetRole updates the role of the user with the given id.
	SetRole(ctx context.Context, id uint, role model.Role) error
	// Delete removes the user with the given id.
	Delete(ctx context.Context, id uint) error
}
