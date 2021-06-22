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
	"github.com/1995parham/koochooloo/internal/metric"
	"github.com/1995parham/koochooloo/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main(cfg config.Config, logger *zap.Logger, tracer trace.Tracer) {
	metric.NewServer(cfg.Monitoring).Start(logger.Named("metrics"))

	app := echo.New()

	db, err := db.New(cfg.Database)
	if err != nil {
		logger.Fatal("database initiation failed", zap.Error(err))
	}

	handler.URL{
		Store:  store.NewMongoURL(db, tracer),
		Logger: logger.Named("handler").Named("url"),
		Tracer: tracer,
	}.Register(app.Group("/api"))

	handler.Healthz{
		Logger: logger.Named("handler").Named("healthz"),
		Tracer: tracer,
	}.Register(app.Group(""))

	if err := app.Start(":1378"); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("echo initiation failed", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Register server command.
func Register(root *cobra.Command, cfg config.Config, logger *zap.Logger, tracer trace.Tracer) {
	root.AddCommand(
		// nolint: exhaustivestruct
		&cobra.Command{
			Use:   "server",
			Short: "Run server to serve the requests",
			Run: func(cmd *cobra.Command, args []string) {
				main(cfg, logger, tracer)
			},
		},
	)
}
