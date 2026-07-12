package model

import "time"

// Role is a user's privilege tier. Roles are ordered: user < admin < superadmin.
type Role string

const (
	// RoleUser can manage only their own short URLs.
	RoleUser Role = "user"
	// RoleAdmin can manage every short URL and view users.
	RoleAdmin Role = "admin"
	// RoleSuperAdmin can additionally manage users and their roles.
	RoleSuperAdmin Role = "superadmin"
)

// privilege ranks, ordered least to most privileged. rankUnknown is the zero
// value so an unrecognised role never satisfies AtLeast.
const (
	rankUnknown = iota
	rankUser
	rankAdmin
	rankSuperAdmin
)

// Valid reports whether the role is one of the known roles.
func (r Role) Valid() bool {
	return r.rank() > rankUnknown
}

// AtLeast reports whether the role is at least as privileged as other.
func (r Role) AtLeast(other Role) bool {
	return r.rank() >= other.rank()
}

// rank returns the privilege level of the role; higher is more privileged.
func (r Role) rank() int {
	switch r {
	case RoleUser:
		return rankUser
	case RoleAdmin:
		return rankAdmin
	case RoleSuperAdmin:
		return rankSuperAdmin
	default:
		return rankUnknown
	}
}

// Provider identifies how a user authenticates.
type Provider string

const (
	// ProviderLocal is a username/password account managed by koochooloo.
	ProviderLocal Provider = "local"
	// ProviderOIDC is an account federated from an external OIDC provider.
	ProviderOIDC Provider = "oidc"
)

// User is an account that can sign in and own short URLs.
type User struct {
	ID           uint
	Username     string
	PasswordHash string // empty for OIDC-only accounts
	Role         Role
	Provider     Provider
	Subject      string // OIDC subject ("sub" claim); empty for local accounts
	CreatedAt    time.Time
}
