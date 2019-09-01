package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/1995parham/url-shortener/config"
	"gitlab.com/1995parham/url-shortener/db"
	"gitlab.com/1995parham/url-shortener/handlers"
	"gitlab.com/1995parham/url-shortener/router"
	"gitlab.com/1995parham/url-shortener/stores"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()

	e := router.App(cfg.Debug)

	// routes
	db, err := db.New(cfg.Database.URL, "urlshortener")
	if err != nil {
		logrus.Fatal(err)
	}

	uh := handlers.URLHandler{
		Store: stores.URLStore{
			DB: db,
		},
	}

	api := e.Group("/api")
	{
		uh.Register(api)
	}

	go func() {
		if err := e.Start(":8080"); err != http.ErrServerClosed {
			logrus.Fatalf("API Service failed with %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("API Service failed on exit: %s", err)
	}
}
