package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// URL handles interaction with URLs.
type URL struct {
	Store  *urlsvc.URLSvc
	Logger *zap.Logger
	Tracer trace.Tracer
}

// Create generates short URL and save it on database.
func (h URL) Create(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.create")
	defer span.End()

	var rq request.URL

	if err := c.Bind(&rq); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	span.SetAttributes(attribute.String("url", rq.URL))

	k, err := h.Store.Set(ctx, rq.Name, rq.URL, rq.Expire, 0)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, urlrepo.ErrDuplicateKey) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, k)
}

// Retrieve retrieves URL for given short URL and redirect to it.
func (h URL) Retrieve(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.retrieve")
	defer span.End()

	return h.redirect(ctx, c, span, c.Param("key"))
}

// RootRetrieve resolves a short URL served at the host root (`/<key>`), the
// form a dedicated short-link host hands out. A custom alias is stored with a
// "$" prefix, so `/<name>` is tried as a generated key first and then as the
// alias `$<name>`; `/$<name>` still matches the alias directly.
func (h URL) RootRetrieve(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.root")
	defer span.End()

	key := c.Param("key")

	if !strings.HasPrefix(key, "$") {
		if _, err := h.Store.Get(ctx, key); errors.Is(err, urlrepo.ErrKeyNotFound) {
			key = "$" + key
		}
	}

	return h.redirect(ctx, c, span, key)
}

// Count retrieves the access count for the given short URL.
func (h URL) Count(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.count")
	defer span.End()

	key := c.Param("key")

	count, err := h.Store.Count(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, count)
}

// Register registers the routes of URL handler on given group.
func (h URL) Register(g *echo.Group) {
	g.GET("/:key", h.Retrieve)
	g.POST("/urls", h.Create)
	g.GET("/count/:key", h.Count)
}

// RegisterRoot serves short URLs at the host root. Echo matches static routes
// before parameters, so `/healthz` and `/admin` keep winning over `/:key`.
func (h URL) RegisterRoot(app *echo.Echo) {
	app.GET("/:key", h.RootRetrieve)
}

// redirect resolves key, counts the visit and answers with a 302 to its target.
func (h URL) redirect(ctx context.Context, c *echo.Context, span trace.Span, key string) error {
	u, err := h.Store.ResolveAndTrack(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, urlrepo.ErrKeyNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		h.Logger.Error("resolve and track failed",
			zap.Error(err),
			zap.String("key", key),
		)
	}

	if u.URL == "" {
		return echo.NewHTTPError(http.StatusNotFound, "url not found")
	}

	return c.Redirect(http.StatusFound, u.URL)
}
