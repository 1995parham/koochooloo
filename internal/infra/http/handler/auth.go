package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/service/usersvc"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/1995parham/koochooloo/internal/infra/http/response"
	"github.com/1995parham/koochooloo/internal/infra/oidc"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Auth handles authentication: local login, the current-user endpoint and the
// OIDC authorization-code flow.
type Auth struct {
	Users    *usersvc.UserSvc
	Tokens   *auth.TokenService
	Provider *oidc.Service
	Logger   *zap.Logger
	Tracer   trace.Tracer
}

// Login verifies credentials and returns a signed session token.
func (h Auth) Login(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.auth.login")
	defer span.End()

	var rq request.Login

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.Users.Authenticate(ctx, rq.Username, rq.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, usersvc.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
	}

	token, err := h.Tokens.Issue(user, time.Now())
	if err != nil {
		span.RecordError(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "issuing token failed")
	}

	return c.JSON(http.StatusOK, response.Token{Token: token, User: response.NewUser(user)})
}

// Me returns the currently authenticated user.
func (h Auth) Me(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.auth.me")
	defer span.End()

	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	}

	id, err := claims.UserID()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
	}

	user, err := h.Users.Get(ctx, id)
	if err != nil {
		span.RecordError(err)

		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, response.NewUser(user))
}
