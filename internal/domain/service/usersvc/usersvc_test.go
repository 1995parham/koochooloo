package usersvc_test

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/usersvc"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"github.com/stretchr/testify/require"
)

func newSvc() *usersvc.UserSvc {
	return usersvc.Provide(userdb.ProvideMemory())
}

func TestRegisterAndAuthenticate(t *testing.T) {
	t.Parallel()

	svc := newSvc()
	ctx := t.Context()

	user, err := svc.Register(ctx, "raha", "s3cret", model.RoleAdmin)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	require.Equal(t, model.RoleAdmin, user.Role)
	require.Equal(t, model.ProviderLocal, user.Provider)
	require.NotEqual(t, "s3cret", user.PasswordHash, "password must be hashed")

	t.Run("correct password", func(t *testing.T) {
		t.Parallel()

		got, err := svc.Authenticate(ctx, "raha", "s3cret")
		require.NoError(t, err)
		require.Equal(t, user.ID, got.ID)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		_, err := svc.Authenticate(ctx, "raha", "nope")
		require.ErrorIs(t, err, usersvc.ErrInvalidCredentials)
	})

	t.Run("unknown user", func(t *testing.T) {
		t.Parallel()

		_, err := svc.Authenticate(ctx, "ghost", "whatever")
		require.ErrorIs(t, err, usersvc.ErrInvalidCredentials)
	})
}

func TestRegisterInvalidRole(t *testing.T) {
	t.Parallel()

	_, err := newSvc().Register(t.Context(), "x", "y", model.Role("root"))
	require.ErrorIs(t, err, usersvc.ErrInvalidRole)
}

func TestEnsureOIDC(t *testing.T) {
	t.Parallel()

	svc := newSvc()
	ctx := t.Context()

	// first login provisions the account
	user, err := svc.EnsureOIDC(ctx, "sub-1", "keycloak-user", model.RoleAdmin)
	require.NoError(t, err)
	require.Equal(t, model.ProviderOIDC, user.Provider)
	require.Equal(t, model.RoleAdmin, user.Role)

	// OIDC accounts cannot authenticate with a password
	_, err = svc.Authenticate(ctx, "keycloak-user", "")
	require.ErrorIs(t, err, usersvc.ErrInvalidCredentials)

	// second login is idempotent and reflects an external role change
	again, err := svc.EnsureOIDC(ctx, "sub-1", "keycloak-user", model.RoleSuperAdmin)
	require.NoError(t, err)
	require.Equal(t, user.ID, again.ID)
	require.Equal(t, model.RoleSuperAdmin, again.Role)
}

func TestSetRoleAndDelete(t *testing.T) {
	t.Parallel()

	svc := newSvc()
	ctx := t.Context()

	user, err := svc.Register(ctx, "u", "p", model.RoleUser)
	require.NoError(t, err)

	require.NoError(t, svc.SetRole(ctx, user.ID, model.RoleAdmin))
	require.ErrorIs(t, svc.SetRole(ctx, user.ID, model.Role("bad")), usersvc.ErrInvalidRole)

	users, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, users, 1)

	require.NoError(t, svc.Delete(ctx, user.ID))
	require.ErrorIs(t, svc.Delete(ctx, user.ID), userrepo.ErrUserNotFound)
}
