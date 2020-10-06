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
)

func main(cfg config.Config) {
	metric.NewServer(cfg.Monitoring).Start()

	app := echo.New()

	db, err := db.New(cfg.Database.URL, cfg.Database.Name)
	if err != nil {
		panic(err)
	}

	handler.URL{
		Store: store.NewMongoURL(db),
	}.Register(app.Group("/api"))

	handler.Healthz{}.Register(app.Group("/"))

	if err := app.Start(":1378"); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Register server command.
func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "server",
			Short: "Run server to serve the requests",
			Run: func(cmd *cobra.Command, args []string) {
				main(cfg)
			},
		},
	)
}
