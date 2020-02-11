package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/1995parham/koochooloo/config"
	"github.com/1995parham/koochooloo/db"
	"github.com/1995parham/koochooloo/handler"
	"github.com/1995parham/koochooloo/store"
	"github.com/gofiber/fiber"
	"github.com/spf13/cobra"
)

func main(cfg config.Config) {
	app := fiber.New()
	app.Prefork = true // Prefork enabled

	db, err := db.New(cfg.Database.URL, "urlshortener")
	if err != nil {
		panic(err)
	}

	handler.URLHandler{
		Store: store.URL{DB: db},
	}.Register(app.Group("/api"))

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
