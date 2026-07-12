package oidc

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/stretchr/testify/require"
)

const rolesKey = "roles"

func keycloakClaims(roles ...any) map[string]any {
	return map[string]any{
		"preferred_username": "raha",
		"realm_access":       map[string]any{rolesKey: roles},
	}
}

func mapper() *Service {
	return &Service{ //nolint:exhaustruct
		cfg: auth.OIDCConfig{ //nolint:exhaustruct
			RolesClaim:       "realm_access.roles",
			AdminValues:      []string{"kc-admin"},
			SuperAdminValues: []string{"kc-superadmin"},
		},
	}
}

func TestMapRole(t *testing.T) {
	t.Parallel()

	svc := mapper()

	require.Equal(t, model.RoleSuperAdmin, svc.mapRole(keycloakClaims("offline_access", "kc-superadmin")))
	require.Equal(t, model.RoleAdmin, svc.mapRole(keycloakClaims("kc-admin")))
	require.Equal(t, model.RoleUser, svc.mapRole(keycloakClaims("something-else")))
	require.Equal(t, model.RoleUser, svc.mapRole(map[string]any{"no": "roles"}))

	// superadmin wins even when both are present
	require.Equal(t, model.RoleSuperAdmin, svc.mapRole(keycloakClaims("kc-admin", "kc-superadmin")))
}

func TestMapRoleNoClaimConfigured(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct
	svc := &Service{cfg: auth.OIDCConfig{RolesClaim: ""}}
	require.Equal(t, model.RoleUser, svc.mapRole(keycloakClaims("kc-superadmin")))
}

func TestClaimValues(t *testing.T) {
	t.Parallel()

	claims := map[string]any{
		"groups":       []any{"a", "b", 42}, // non-strings are skipped
		"realm_access": map[string]any{rolesKey: []any{"x"}},
		"scalar":       "single",
	}

	require.ElementsMatch(t, []string{"a", "b"}, claimValues(claims, "groups"))
	require.Equal(t, []string{"x"}, claimValues(claims, "realm_access.roles"))
	require.Equal(t, []string{"single"}, claimValues(claims, "scalar"))
	require.Nil(t, claimValues(claims, "missing.path"))
}

func TestDisabledService(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct
	svc := &Service{enabled: false}
	require.False(t, svc.Enabled())

	_, err := svc.Verify(t.Context(), "code", "nonce")
	require.ErrorIs(t, err, ErrDisabled)
}
