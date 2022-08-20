package server

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1995parham/koochooloo/internal"
	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/handler"
	"github.com/1995parham/koochooloo/pkg/telemetry"
	"github.com/1995parham/koochooloo/pkg/telemetry/log"
	"github.com/1995parham/koochooloo/pkg/telemetry/metric"
	"github.com/1995parham/koochooloo/pkg/telemetry/trace"

	// "github.com/1995parham/koochooloo/internal/metric"
	"github.com/1995parham/koochooloo/internal/store/url"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main(cfg *config.Config) {
	telemetry := &telemetry.Telemetry{
		Log:    log.NewZap(cfg.Telemetry.Log),
		Metric: metric.New(internal.Namespace, internal.Subsystem),
		Trace:  trace.New(cfg.Telemetry.Trace, internal.Namespace, internal.Subsystem),
	}

	if err := metric.NewServer(cfg.Telemetry.Metric).Serve(); err != nil {
		telemetry.Log.Fatal("metric serving failed", zap.Error(err))
	}

	app := echo.New()

	db, err := db.New(cfg.Database)
	if err != nil {
		telemetry.Log.Fatal("database initiation failed", zap.Error(err))
	}

	handler.URL{
		Store:  url.NewMongoURL(db, telemetry.Trace),
		Logger: telemetry.Log.Named("handler").Named("url"),
		Tracer: telemetry.Trace,
	}.Register(app.Group("/api"))

	handler.Healthz{
		Logger: telemetry.Log.Named("handler").Named("healthz"),
		Tracer: telemetry.Trace,
	}.Register(app.Group(""))

	if err := app.Start(":1378"); !errors.Is(err, http.ErrServerClosed) {
		telemetry.Log.Fatal("echo initiation failed", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Register server command.
func Register(root *cobra.Command, cfg *config.Config) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "server",
			Short: "Run server to serve the requests",
			Run:   func(cmd *cobra.Command, args []string) { main(cfg) },
		},
	)
}
