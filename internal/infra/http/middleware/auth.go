// Package middleware provides echo middleware for the admin API: JWT and
// API-key authentication and role-based authorization.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/service/apikeysvc"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/labstack/echo/v5"
)

// claimsContextKey is the echo context key under which the verified token
// claims are stored for downstream handlers.
const claimsContextKey = "koochooloo.claims"

// KeyAuthenticator resolves an API key to the user that owns it.
type KeyAuthenticator interface {
	Authenticate(ctx context.Context, key string) (model.User, error)
}

// Auth guards routes with a bearer session JWT or a bearer API key.
type Auth struct {
	Tokens *auth.TokenService
	// Keys resolves `kch_` API keys; nil rejects every key.
	Keys KeyAuthenticator
}

// Authenticate verifies the Authorization bearer token and stores the claims
// in the request context. A value starting with the API-key prefix is checked
// as an API key and acts as the key's owner; anything else must be a session
// JWT. Requests without a valid credential are rejected with 401.
func (a Auth) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		header := c.Request().Header.Get("Authorization")

		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}

		if strings.HasPrefix(token, apikeysvc.Prefix) {
			return a.authenticateKey(c, next, token)
		}

		claims, err := a.Tokens.Parse(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		}

		c.Set(claimsContextKey, claims)

		return next(c)
	}
}

func (a Auth) authenticateKey(c *echo.Context, next echo.HandlerFunc, key string) error {
	if a.Keys == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
	}

	user, err := a.Keys.Authenticate(c.Request().Context(), key)
	if err != nil {
		if errors.Is(err, apikeysvc.ErrInvalidKey) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "checking api key failed")
	}

	c.Set(claimsContextKey, auth.ClaimsFor(user))

	return next(c)
}

// RequireRole rejects requests whose authenticated user is below the given
// role with 403. It must run after Authenticate.
func RequireRole(minimum model.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, ok := ClaimsFrom(c)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
			}

			if !claims.Role.AtLeast(minimum) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient privileges")
			}

			return next(c)
		}
	}
}

// ClaimsFrom returns the verified claims stored by Authenticate.
func ClaimsFrom(c *echo.Context) (auth.Claims, bool) {
	claims, ok := c.Get(claimsContextKey).(auth.Claims)

	return claims, ok
}
