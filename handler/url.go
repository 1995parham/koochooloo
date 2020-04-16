package handler

import (
	"errors"
	"net/http"

	"github.com/1995parham/koochooloo/request"
	"github.com/1995parham/koochooloo/store"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// URL handles interaction with URLs
type URL struct {
	Store store.URL
}

// Create generates short URL and save it on database
func (h URL) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var rq request.URL

	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := rq.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	k, err := h.Store.Set(ctx, rq.Name, rq.URL, rq.Expire)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, k)
}

// Retrieve retrieves URL for given short URL and redirect to it
func (h URL) Retrieve(c echo.Context) error {
	ctx := c.Request().Context()

	key := c.Param("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := h.Store.Inc(ctx, key); err != nil {
		logrus.Errorf("Inc Error: %s", err)
	}

	return c.Redirect(http.StatusFound, url)
}

// Register registers the routes of URL handler on given group
func (h URL) Register(g *echo.Group) {
	g.GET("/:key", h.Retrieve)
	g.POST("/urls", h.Create)
}
