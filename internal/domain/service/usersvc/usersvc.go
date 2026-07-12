// Package usersvc contains user-account business logic: registration,
// password authentication, OIDC provisioning and role management.
package usersvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials indicates a failed username/password login.
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrInvalidRole indicates an unknown role was supplied.
	ErrInvalidRole = errors.New("invalid role")
)

// UserSvc is the user-account service.
type UserSvc struct {
	repo userrepo.Repository
}

// Provide creates a new user service.
func Provide(repo userrepo.Repository) *UserSvc {
	return &UserSvc{repo: repo}
}

// Register creates a new local (password) account with the given role.
func (s *UserSvc) Register(
	ctx context.Context, username, password string, role model.Role,
) (model.User, error) {
	if !role.Valid() {
		return model.User{}, ErrInvalidRole
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hashing password failed: %w", err)
	}

	user, err := s.repo.Create(ctx, model.User{ //nolint:exhaustruct // ID/CreatedAt assigned by the store.
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Provider:     model.ProviderLocal,
	})
	if err != nil {
		return model.User{}, fmt.Errorf("creating user failed: %w", err)
	}

	return user, nil
}

// Authenticate verifies a username/password pair and returns the user.
func (s *UserSvc) Authenticate(ctx context.Context, username, password string) (model.User, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return model.User{}, ErrInvalidCredentials
		}

		return model.User{}, fmt.Errorf("looking up user failed: %w", err)
	}

	// OIDC-only accounts have no password and cannot log in locally.
	if user.PasswordHash == "" {
		return model.User{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, ErrInvalidCredentials
	}

	return user, nil
}

// EnsureOIDC returns the account federated from the OIDC provider with the
// given subject, provisioning it on first sight (just-in-time). The role is
// applied on every login so external role changes are reflected.
func (s *UserSvc) EnsureOIDC(
	ctx context.Context, subject, username string, role model.Role,
) (model.User, error) {
	if !role.Valid() {
		role = model.RoleUser
	}

	user, err := s.repo.FindBySubject(ctx, model.ProviderOIDC, subject)
	switch {
	case err == nil:
		if user.Role != role {
			if err := s.repo.SetRole(ctx, user.ID, role); err != nil {
				return model.User{}, fmt.Errorf("updating oidc role failed: %w", err)
			}

			user.Role = role
		}

		return user, nil
	case errors.Is(err, userrepo.ErrUserNotFound):
		created, err := s.repo.Create(ctx, model.User{ //nolint:exhaustruct // no password for OIDC accounts.
			Username: username,
			Role:     role,
			Provider: model.ProviderOIDC,
			Subject:  subject,
		})
		if err != nil {
			return model.User{}, fmt.Errorf("provisioning oidc user failed: %w", err)
		}

		return created, nil
	default:
		return model.User{}, fmt.Errorf("looking up oidc user failed: %w", err)
	}
}

// Get returns the user with the given id.
func (s *UserSvc) Get(ctx context.Context, id uint) (model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.User{}, fmt.Errorf("looking up user failed: %w", err)
	}

	return user, nil
}

// List returns all users.
func (s *UserSvc) List(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users failed: %w", err)
	}

	return users, nil
}

// SetRole changes a user's role.
func (s *UserSvc) SetRole(ctx context.Context, id uint, role model.Role) error {
	if !role.Valid() {
		return ErrInvalidRole
	}

	if err := s.repo.SetRole(ctx, id, role); err != nil {
		return fmt.Errorf("setting role failed: %w", err)
	}

	return nil
}

// Delete removes a user.
func (s *UserSvc) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting user failed: %w", err)
	}

	return nil
}
