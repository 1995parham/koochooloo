package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/http/handler"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	port               = ":1378"
	readHeaderTimeout  = 10 * time.Second
)

func Provide(lc fx.Lifecycle, store *urlsvc.URLSvc, logger *zap.Logger, tele telemetry.Telemetery) *echo.Echo {
	app := echo.New()

	handler.URL{
		Store:  store,
		Logger: logger.Named("handler").Named("url"),
		Tracer: tele.TraceProvider.Tracer("handler.url"),
	}.Register(app.Group("/api"))

	handler.Healthz{
		Logger: logger.Named("handler").Named("healthz"),
		Tracer: tele.TraceProvider.Tracer("handler.healthz"),
	}.Register(app.Group(""))

	//nolint: exhaustruct
	srv := &http.Server{
		Addr:              port,
		Handler:           app,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
						logger.Fatal("echo initiation failed", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: srv.Shutdown,
		},
	)

	return app
}
