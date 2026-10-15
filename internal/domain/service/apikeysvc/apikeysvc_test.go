package apikeysvc_test

import (
	"strings"
	"testing"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/apikeysvc"
	"github.com/1995parham/koochooloo/internal/infra/repository/apikeydb"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"github.com/stretchr/testify/require"
)

// setup returns a service over in-memory stores holding one user, "svc".
func setup(t *testing.T) (*apikeysvc.APIKeySvc, *apikeydb.MemoryAPIKey, model.User) {
	t.Helper()

	users := userdb.ProvideMemory()
	keys := apikeydb.ProvideMemory()

	user, err := users.Create(t.Context(), model.User{ //nolint:exhaustruct_v5 // ID/CreatedAt assigned by the store.
		Username: "svc",
		Role:     model.RoleUser,
		Provider: model.ProviderLocal,
	})
	require.NoError(t, err)

	return apikeysvc.Provide(keys, users), keys, user
}

func TestCreateAndAuthenticate(t *testing.T) {
	t.Parallel()

	svc, keys, user := setup(t)
	ctx := t.Context()

	plain, key, err := svc.Create(ctx, "svc", "hatchup-team")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(plain, apikeysvc.Prefix))
	require.Len(t, plain, len(apikeysvc.Prefix)+40)
	require.Equal(t, user.ID, key.UserID)
	require.Equal(t, "hatchup-team", key.Name)
	require.True(t, strings.HasPrefix(plain, key.Prefix))
	require.NotContains(t, key.Hash, plain, "only a hash of the key is stored")

	got, err := svc.Authenticate(ctx, plain)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, model.RoleUser, got.Role)

	stored, err := keys.FindByHash(ctx, key.Hash)
	require.NoError(t, err)
	require.NotNil(t, stored.LastUsedAt, "a successful use is recorded")

	other, _, err := svc.Create(ctx, "svc", "second")
	require.NoError(t, err)
	require.NotEqual(t, plain, other, "every key is random")
}

func TestAuthenticateRejects(t *testing.T) {
	t.Parallel()

	svc, _, _ := setup(t)
	ctx := t.Context()

	plain, key, err := svc.Create(ctx, "svc", "hatchup-team")
	require.NoError(t, err)

	cases := map[string]string{
		"no prefix":   strings.TrimPrefix(plain, apikeysvc.Prefix),
		"too short":   plain[:len(plain)-1],
		"unknown key": apikeysvc.Prefix + strings.Repeat("a", 40),
		"empty":       "",
	}

	for name, candidate := range cases {
		_, err := svc.Authenticate(ctx, candidate)
		require.ErrorIs(t, err, apikeysvc.ErrInvalidKey, name)
	}

	require.NoError(t, svc.Revoke(ctx, key.ID))

	_, err = svc.Authenticate(ctx, plain)
	require.ErrorIs(t, err, apikeysvc.ErrInvalidKey, "a revoked key stops working")

	listed, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.True(t, listed[0].Revoked())
}

func TestCreateValidation(t *testing.T) {
	t.Parallel()

	svc, _, _ := setup(t)

	_, _, err := svc.Create(t.Context(), "svc", "  ")
	require.ErrorIs(t, err, apikeysvc.ErrEmptyName)

	_, _, err = svc.Create(t.Context(), "ghost", "x")
	require.ErrorIs(t, err, userrepo.ErrUserNotFound)
}

func TestRevokeUnknown(t *testing.T) {
	t.Parallel()

	svc, _, _ := setup(t)

	require.Error(t, svc.Revoke(t.Context(), 99))
}
