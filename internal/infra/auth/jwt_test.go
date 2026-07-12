package auth_test

import (
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/stretchr/testify/require"
)

func user() model.User {
	return model.User{ //nolint:exhaustruct
		ID:       42,
		Username: "raha",
		Role:     model.RoleSuperAdmin,
	}
}

func TestIssueAndParse(t *testing.T) {
	t.Parallel()

	svc := auth.NewTokenService("secret", time.Hour)

	token, err := svc.Issue(user(), time.Now())
	require.NoError(t, err)

	claims, err := svc.Parse(token)
	require.NoError(t, err)
	require.Equal(t, "raha", claims.Username)
	require.Equal(t, model.RoleSuperAdmin, claims.Role)

	id, err := claims.UserID()
	require.NoError(t, err)
	require.Equal(t, uint(42), id)
}

func TestParseExpired(t *testing.T) {
	t.Parallel()

	svc := auth.NewTokenService("secret", time.Hour)

	// issued two hours ago with a one-hour TTL -> expired
	token, err := svc.Issue(user(), time.Now().Add(-2*time.Hour))
	require.NoError(t, err)

	_, err = svc.Parse(token)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseWrongSecret(t *testing.T) {
	t.Parallel()

	token, err := auth.NewTokenService("secret", time.Hour).Issue(user(), time.Now())
	require.NoError(t, err)

	_, err = auth.NewTokenService("other-secret", time.Hour).Parse(token)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseGarbage(t *testing.T) {
	t.Parallel()

	_, err := auth.NewTokenService("secret", time.Hour).Parse("not-a-jwt")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}
