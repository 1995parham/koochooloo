package handler

import (
	"errors"
	"net/http"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/1995parham/koochooloo/internal/infra/http/response"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// AdminURL handles admin-panel short URL management. Regular users see and
// manage only their own shorts; admins and superadmins manage all of them.
type AdminURL struct {
	Store  *urlsvc.URLSvc
	Logger *zap.Logger
	Tracer trace.Tracer
}

// List returns the caller's short URLs, or every short URL for admins.
func (h AdminURL) List(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminurl.list")
	defer span.End()

	claims, id, err := caller(c)
	if err != nil {
		return err
	}

	var urls []model.URL

	if claims.Role.AtLeast(model.RoleAdmin) {
		urls, err = h.Store.ListAll(ctx)
	} else {
		urls, err = h.Store.ListByOwner(ctx, id)
	}

	if err != nil {
		span.RecordError(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "listing urls failed")
	}

	return c.JSON(http.StatusOK, response.URLs(urls))
}

// Create makes a new short URL owned by the caller.
func (h AdminURL) Create(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminurl.create")
	defer span.End()

	_, id, err := caller(c)
	if err != nil {
		return err
	}

	var rq request.URL

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	key, err := h.Store.SetForOwner(ctx, rq.Name, rq.URL, rq.Expire, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, urlrepo.ErrDuplicateKey) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "creating url failed")
	}

	return c.JSON(http.StatusCreated, map[string]string{"key": key})
}

// Delete removes a short URL. Regular users may only delete their own.
func (h AdminURL) Delete(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminurl.delete")
	defer span.End()

	claims, id, err := caller(c)
	if err != nil {
		return err
	}

	key := c.Param("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, urlrepo.ErrKeyNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "url not found")
		}

		span.RecordError(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "looking up url failed")
	}

	owned := url.OwnerID != nil && *url.OwnerID == id
	if !claims.Role.AtLeast(model.RoleAdmin) && !owned {
		return echo.NewHTTPError(http.StatusForbidden, "not your url")
	}

	if err := h.Store.Delete(ctx, key); err != nil {
		span.RecordError(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "deleting url failed")
	}

	return c.NoContent(http.StatusNoContent)
}

// caller returns the authenticated claims and numeric user id, or an HTTP error.
func caller(c *echo.Context) (auth.Claims, uint, error) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		return auth.Claims{}, 0, echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	}

	id, err := claims.UserID()
	if err != nil {
		return auth.Claims{}, 0, echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
	}

	return claims, id, nil
}
