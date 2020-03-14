package server

import (
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
	metric.NewServer(cfg.Monitoring)

	app := echo.New()

	db, err := db.New(cfg.Database.URL, cfg.Database.Name)
	if err != nil {
		panic(err)
	}

	handler.URLHandler{
		Store: store.NewURL(db),
	}.Register(app.Group("/api"))

	if err := app.Start(":1378"); err != http.ErrServerClosed {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// Register server command
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
