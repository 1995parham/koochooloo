package server

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1995parham/koochooloo/config"
	"github.com/1995parham/koochooloo/db"
	"github.com/1995parham/koochooloo/handler"
	"github.com/1995parham/koochooloo/metric"
	"github.com/1995parham/koochooloo/store"
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

	app.GET("/healthz", func(context echo.Context) error {
		return context.NoContent(http.StatusNoContent)
	})

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
