package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/usersvc"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/1995parham/koochooloo/internal/infra/http/response"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// AdminUser handles user-account management. Route middleware enforces the
// minimum role (admin to list, superadmin to mutate).
type AdminUser struct {
	Users  *usersvc.UserSvc
	Logger *zap.Logger
	Tracer trace.Tracer
}

// List returns all user accounts.
func (h AdminUser) List(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminuser.list")
	defer span.End()

	users, err := h.Users.List(ctx)
	if err != nil {
		span.RecordError(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "listing users failed")
	}

	return c.JSON(http.StatusOK, response.Users(users))
}

// Create makes a new local user account.
func (h AdminUser) Create(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminuser.create")
	defer span.End()

	var rq request.CreateUser

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.Users.Register(ctx, rq.Username, rq.Password, rq.Role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, userrepo.ErrDuplicateUsername) {
			return echo.NewHTTPError(http.StatusConflict, "username already exists")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "creating user failed")
	}

	return c.JSON(http.StatusCreated, response.NewUser(user))
}

// SetRole changes a user's role.
func (h AdminUser) SetRole(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminuser.set_role")
	defer span.End()

	id, err := pathID(c)
	if err != nil {
		return err
	}

	var rq request.SetRole

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.Users.SetRole(ctx, id, rq.Role); err != nil {
		span.RecordError(err)

		if errors.Is(err, userrepo.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "setting role failed")
	}

	return c.NoContent(http.StatusNoContent)
}

// Delete removes a user account.
func (h AdminUser) Delete(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.adminuser.delete")
	defer span.End()

	id, err := pathID(c)
	if err != nil {
		return err
	}

	if err := h.Users.Delete(ctx, id); err != nil {
		span.RecordError(err)

		if errors.Is(err, userrepo.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "deleting user failed")
	}

	return c.NoContent(http.StatusNoContent)
}

// pathID parses the :id path parameter as a user id.
func pathID(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	return uint(id), nil
}
