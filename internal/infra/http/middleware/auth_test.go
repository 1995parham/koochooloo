package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func do(engine *echo.Echo, path, authz string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	if authz != "" {
		req.Header.Set("Authorization", authz)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func token(t *testing.T, tokens *auth.TokenService, role model.Role) string {
	t.Helper()

	signed, err := tokens.Issue(model.User{ //nolint:exhaustruct
		ID:       7,
		Username: "u",
		Role:     role,
	}, time.Now())
	require.NoError(t, err)

	return signed
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	engine := echo.New()
	tokens := auth.NewTokenService("secret", time.Hour)
	mw := middleware.Auth{Tokens: tokens}

	group := engine.Group("", mw.Authenticate)
	group.GET("/x", func(c *echo.Context) error {
		claims, ok := middleware.ClaimsFrom(c)
		require.True(t, ok)

		return c.String(http.StatusOK, string(claims.Role))
	})

	require.Equal(t, http.StatusOK, do(engine, "/x", "Bearer "+token(t, tokens, model.RoleUser)).Code)
	require.Equal(t, http.StatusUnauthorized, do(engine, "/x", "").Code)
	require.Equal(t, http.StatusUnauthorized, do(engine, "/x", "Bearer garbage").Code)
	require.Equal(t, http.StatusUnauthorized, do(engine, "/x", token(t, tokens, model.RoleUser)).Code) // no "Bearer "
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	engine := echo.New()
	tokens := auth.NewTokenService("secret", time.Hour)
	mw := middleware.Auth{Tokens: tokens}

	group := engine.Group("", mw.Authenticate)
	group.GET("/admin", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, middleware.RequireRole(model.RoleAdmin))

	require.Equal(t, http.StatusForbidden, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleUser)).Code)
	require.Equal(t, http.StatusOK, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleAdmin)).Code)
	require.Equal(t, http.StatusOK, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleSuperAdmin)).Code)
}
