package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/service/apikeysvc"
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

	signed, err := tokens.Issue(model.User{ //nolint:exhaustruct_v5
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
	mw := middleware.Auth{Tokens: tokens, Keys: nil}

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
	mw := middleware.Auth{Tokens: tokens, Keys: nil}

	group := engine.Group("", mw.Authenticate)
	group.GET("/admin", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, middleware.RequireRole(model.RoleAdmin))

	require.Equal(t, http.StatusForbidden, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleUser)).Code)
	require.Equal(t, http.StatusOK, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleAdmin)).Code)
	require.Equal(t, http.StatusOK, do(engine, "/admin", "Bearer "+token(t, tokens, model.RoleSuperAdmin)).Code)
}

// keys is a fake KeyAuthenticator holding one valid key.
type keys struct {
	valid string
	user  model.User
}

func (k keys) Authenticate(_ context.Context, key string) (model.User, error) {
	if key != k.valid {
		return model.User{}, apikeysvc.ErrInvalidKey
	}

	return k.user, nil
}

func TestAuthenticateAPIKey(t *testing.T) {
	t.Parallel()

	valid := apikeysvc.Prefix + strings.Repeat("k", 40)
	owner := model.User{ //nolint:exhaustruct_v5 // only the identity fields matter.
		ID:       9,
		Username: "hatchup-team",
		Role:     model.RoleUser,
	}

	engine := echo.New()
	tokens := auth.NewTokenService("secret", time.Hour)
	mw := middleware.Auth{Tokens: tokens, Keys: keys{valid: valid, user: owner}}

	group := engine.Group("", mw.Authenticate)
	group.GET("/x", func(c *echo.Context) error {
		claims, ok := middleware.ClaimsFrom(c)
		require.True(t, ok)

		id, err := claims.UserID()
		require.NoError(t, err)

		return c.String(http.StatusOK, claims.Username+"/"+strconv.FormatUint(uint64(id), 10))
	})
	group.GET("/admin", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, middleware.RequireRole(model.RoleAdmin))

	ok := do(engine, "/x", "Bearer "+valid)
	require.Equal(t, http.StatusOK, ok.Code)
	require.Equal(t, "hatchup-team/9", ok.Body.String())

	require.Equal(t, http.StatusUnauthorized, do(engine, "/x", "Bearer "+apikeysvc.Prefix+"wrong").Code)
	require.Equal(t, http.StatusForbidden, do(engine, "/admin", "Bearer "+valid).Code, "a key acts with its owner's role")

	// Session tokens keep working beside keys.
	require.Equal(t, http.StatusOK, do(engine, "/x", "Bearer "+token(t, tokens, model.RoleUser)).Code)
}

func TestAuthenticateAPIKeyWithoutKeyStore(t *testing.T) {
	t.Parallel()

	engine := echo.New()
	mw := middleware.Auth{Tokens: auth.NewTokenService("secret", time.Hour), Keys: nil}

	engine.Group("", mw.Authenticate).GET("/x", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	require.Equal(t, http.StatusUnauthorized, do(engine, "/x", "Bearer "+apikeysvc.Prefix+strings.Repeat("k", 40)).Code)
}
