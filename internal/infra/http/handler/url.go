package handler

import (
	"errors"
	"net/http"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/labstack/echo/v4"
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
func (h URL) Create(c echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.create")
	defer span.End()

	var rq request.URL

	err := c.Bind(&rq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = rq.Validate()
	if err != nil {
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
func (h URL) Retrieve(c echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.retrieve")
	defer span.End()

	key := c.Param("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	err = h.Store.Inc(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		h.Logger.Error("increase counter for fetching url failed",
			zap.Error(err),
			zap.String("key", key),
			zap.String("url", url),
		)
	}

	return c.Redirect(http.StatusFound, url)
}

// Count retrieves the access count for the given short URL.
func (h URL) Count(c echo.Context) error {
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
