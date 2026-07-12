// Package middleware provides echo middleware for the admin API: JWT
// authentication and role-based authorization.
package middleware

import (
	"net/http"
	"strings"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/labstack/echo/v5"
)

// claimsContextKey is the echo context key under which the verified token
// claims are stored for downstream handlers.
const claimsContextKey = "koochooloo.claims"

// Auth guards routes with a bearer JWT.
type Auth struct {
	Tokens *auth.TokenService
}

// Authenticate verifies the Authorization bearer token and stores the claims
// in the request context. Requests without a valid token are rejected with 401.
func (a Auth) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		header := c.Request().Header.Get("Authorization")

		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}

		claims, err := a.Tokens.Parse(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		}

		c.Set(claimsContextKey, claims)

		return next(c)
	}
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
