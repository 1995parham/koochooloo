package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1995parham/koochooloo/key"
	"github.com/1995parham/koochooloo/request"
	"github.com/1995parham/koochooloo/response"
	"github.com/1995parham/koochooloo/store"
	"github.com/gofiber/fiber"
)

// URLHandler handles interaction with URLs
type URLHandler struct {
	Store store.URLStore
}

// Create generates short URL and save it on database
func (h URLHandler) Create(c *fiber.Ctx) {
	ctx := c.Fasthttp

	var rq request.URL

	if err := json.Unmarshal([]byte(c.Body()), &rq); err != nil {
		if err := c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}); err != nil {
			panic(err)
		}
	}

	if err := rq.Validate(); err != nil {
		if err := c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}); err != nil {
			panic(err)
		}
	}

	var k string
	if rq.Name != "" {
		k = fmt.Sprintf("$%s", rq.Name)

		if err := h.Store.Set(ctx, k, rq.URL, rq.Expire); err != nil {
			if err == store.ErrDuplicateKey {
				if err := c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}); err != nil {
					panic(err)
				}
			}
			if err := c.Status(http.StatusInternalServerError).JSON(response.Error{Message: err.Error()}); err != nil {
				panic(err)
			}
		}
	} else {
		for {
			k = key.Key()

			if err := h.Store.Set(ctx, k, rq.URL, rq.Expire); err != nil {
				if err == store.ErrDuplicateKey {
					continue
				}
				if err := c.Status(http.StatusInternalServerError).JSON(response.Error{Message: err.Error()}); err != nil {
					panic(err)
				}
			}
			break
		}
	}
	if err := c.Status(http.StatusOK).JSON(k); err != nil {
		panic(err)
	}
}

// Retrieve retrieves URL for given short URL and redirect to it
func (h URLHandler) Retrieve(c *fiber.Ctx) {
	ctx := c.Fasthttp

	key := c.Params("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		if err := c.Status(http.StatusNotFound).JSON(response.Error{Message: "Not Found"}); err != nil {
			panic(err)
		}
	}
	if err := h.Store.Inc(ctx, key); err != nil {
		if err := c.Status(http.StatusInternalServerError).JSON(response.Error{Message: err.Error()}); err != nil {
			panic(err)
		}
	}

	c.Status(http.StatusFound).Location(url)
}

// Register registers the routes of URL handler on given echo group
func (h URLHandler) Register(g *fiber.Group) {
	g.Get("/:key", h.Retrieve)
	g.Post("/urls", h.Create)
}
