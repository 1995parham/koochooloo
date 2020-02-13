package handler

import (
	"net/http"

	"github.com/1995parham/koochooloo/request"
	"github.com/1995parham/koochooloo/response"
	"github.com/1995parham/koochooloo/store"
	"github.com/gofiber/fiber"
)

// URLHandler handles interaction with URLs
type URLHandler struct {
	Store store.URL
}

// Create generates short URL and save it on database
func (h URLHandler) Create(c *fiber.Ctx) {
	ctx := c.Fasthttp

	var rq request.URL

	if err := c.BodyParser(&rq); err != nil {
		c.Next(c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}))
	}

	if err := rq.Validate(); err != nil {
		c.Next(c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}))
	}

	k, err := h.Store.Set(ctx, rq.Name, rq.URL, rq.Expire)
	if err != nil {
		if err == store.ErrDuplicateKey {
			c.Next(c.Status(http.StatusBadRequest).JSON(response.Error{Message: err.Error()}))
		}

		c.Next(c.Status(http.StatusInternalServerError).JSON(response.Error{Message: err.Error()}))
	}

	c.Next(c.Status(http.StatusOK).JSON(k))
}

// Retrieve retrieves URL for given short URL and redirect to it
func (h URLHandler) Retrieve(c *fiber.Ctx) {
	ctx := c.Fasthttp

	key := c.Params("key")

	url, err := h.Store.Get(ctx, key)
	if err != nil {
		c.Next(c.Status(http.StatusNotFound).JSON(response.Error{Message: "Not Found"}))
	}

	if err := h.Store.Inc(ctx, key); err != nil {
		c.Next(c.Status(http.StatusInternalServerError).JSON(response.Error{Message: err.Error()}))
	}

	c.Status(http.StatusFound).Location(url)
}

// Register registers the routes of URL handler on given group
func (h URLHandler) Register(g *fiber.Group) {
	g.Get("/:key", h.Retrieve)
	g.Post("/urls", h.Create)
}
