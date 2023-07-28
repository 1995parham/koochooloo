package server

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/handler"
	"github.com/1995parham/koochooloo/internal/store/url"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main(
	cfg config.Config,
	logger *zap.Logger,
) {
	app := echo.New()

	db, err := db.New(cfg.Database)
	if err != nil {
		logger.Fatal("database initiation failed", zap.Error(err))
	}

	handler.URL{
		Store:  url.NewMongoURL(db, otel.GetTracerProvider().Tracer("store.url"), otel.GetMeterProvider().Meter("store.url")),
		Logger: logger.Named("handler").Named("url"),
		Tracer: otel.GetTracerProvider().Tracer("handler.url"),
	}.Register(app.Group("/api"))

	handler.Healthz{
		Logger: logger.Named("handler").Named("healthz"),
		Tracer: otel.GetTracerProvider().Tracer("handler.healthz"),
	}.Register(app.Group(""))

	if err := app.Start(":1378"); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("echo initiation failed", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Register server command.
func Register(
	root *cobra.Command,
) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "server",
			Short: "Run server to serve the requests",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
