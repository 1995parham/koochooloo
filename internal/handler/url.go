package handler

import (
	"errors"
	"net/http"

	"github.com/1995parham/koochooloo/internal/request"
	"github.com/1995parham/koochooloo/internal/store"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// URL handles interaction with URLs.
type URL struct {
	Store  store.URL
	Logger *zap.Logger
}

// Create generates short URL and save it on database.
func (h URL) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var rq request.URL

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	k, err := h.Store.Set(ctx, rq.Name, rq.URL, rq.Expire, 0)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, k)
}

// Retrieve retrieves URL for given short URL and redirect to it.
func (h URL) Retrieve(c echo.Context) error {
	ctx := c.Request().Context()

	key := c.Param("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := h.Store.Inc(ctx, key); err != nil {
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
	ctx := c.Request().Context()

	key := c.Param("key")

	count, err := h.Store.Count(ctx, key)
	if err != nil {
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
