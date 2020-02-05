package handlers

import (
	"fmt"
	"net/http"

	"github.com/1995parham/koochooloo/keys"
	"github.com/1995parham/koochooloo/request"
	"github.com/1995parham/koochooloo/stores"

	"github.com/labstack/echo/v4"
)

// URLHandler handles interaction with URLs
type URLHandler struct {
	Store stores.URLStore
}

// Create generates short URL and save it on database
func (h URLHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var rq request.URLReq
	if err := c.Bind(&rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(rq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var key string
	if rq.Name != "" {
		key = fmt.Sprintf("$%s", rq.Name)

		if err := h.Store.Set(ctx, key, rq.URL, rq.Expire); err != nil {
			if err == stores.ErrDuplicateKey {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		for {
			key = keys.Key()

			if err := h.Store.Set(ctx, key, rq.URL, rq.Expire); err != nil {
				if err == stores.ErrDuplicateKey {
					continue
				}
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			break
		}
	}
	return c.JSON(http.StatusOK, key)
}

// Retrieve retrieves URL for given short URL and redirect to it
func (h URLHandler) Retrieve(c echo.Context) error {
	ctx := c.Request().Context()

	key := c.Param("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusNotFound, "not found")
	}
	if err := h.Store.Inc(ctx, key); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("http://%s", url))
}

// Register registers the routes of URL handler on given echo group
func (h URLHandler) Register(g *echo.Group) {
	g.GET("/:key", h.Retrieve)
	g.POST("/urls", h.Create)
}
