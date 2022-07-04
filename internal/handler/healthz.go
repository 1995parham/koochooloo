package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Healthz struct {
	Logger *zap.Logger
	Tracer trace.Tracer
}

// Handle shows server is up and running.
func (h Healthz) Handle(c echo.Context) error {
	_, span := h.Tracer.Start(c.Request().Context(), "handler.healthz")
	defer span.End()

	return c.NoContent(http.StatusNoContent)
}

// Register registers the routes of healthz handler on given echo group.
func (h Healthz) Register(g *echo.Group) {
	g.GET("/healthz", h.Handle)
}
