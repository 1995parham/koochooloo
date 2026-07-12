package request

import (
	"errors"
	"fmt"

	"github.com/1995parham/koochooloo/internal/domain/model"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// errInvalidRole is returned by the role validator for unknown roles.
var errInvalidRole = errors.New("invalid role")

// Login is the local username/password login payload.
type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate checks that credentials are present.
func (r Login) Validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	); err != nil {
		return fmt.Errorf("login request validation failed: %w", err)
	}

	return nil
}

// CreateUser is the payload for creating a local account (superadmin only).
type CreateUser struct {
	Username string     `json:"username"`
	Password string     `json:"password"`
	Role     model.Role `json:"role"`
}

// Validate checks the new-user payload.
func (r CreateUser) Validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required, validation.Length(1, 255)), //nolint:mnd
		validation.Field(&r.Password, validation.Required, validation.Length(8, 0)),   //nolint:mnd
		validation.Field(&r.Role, validation.Required, validation.By(validRole)),
	); err != nil {
		return fmt.Errorf("create-user request validation failed: %w", err)
	}

	return nil
}

// SetRole is the payload for changing a user's role.
type SetRole struct {
	Role model.Role `json:"role"`
}

// Validate checks the role payload.
func (r SetRole) Validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Role, validation.Required, validation.By(validRole)),
	); err != nil {
		return fmt.Errorf("set-role request validation failed: %w", err)
	}

	return nil
}

// validRole is an ozzo validator that accepts only known roles.
func validRole(value any) error {
	role, ok := value.(model.Role)
	if !ok || !role.Valid() {
		return fmt.Errorf("%w: %q", errInvalidRole, value)
	}

	return nil
}
