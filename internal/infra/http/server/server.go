// Package server wires the echo HTTP server and its route groups.
package server

import (
	"context"
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/domain/service/usersvc"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/1995parham/koochooloo/internal/infra/http/handler"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/1995parham/koochooloo/internal/infra/oidc"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/1995parham/koochooloo/web"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	port              = ":1378"
	readHeaderTimeout = 10 * time.Second
)

// Provide builds the echo server, wiring the public API and the JWT-guarded
// admin API, and manages its lifecycle through fx.
func Provide(
	lc fx.Lifecycle,
	store *urlsvc.URLSvc,
	users *usersvc.UserSvc,
	tokens *auth.TokenService,
	provider *oidc.Service,
	logger *zap.Logger,
	tele telemetry.Telemetery,
) *echo.Echo {
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

	registerAdmin(app, store, users, tokens, provider, logger, tele)
	registerSPA(app)

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

// registerAdmin mounts the /admin/api routes: a public login endpoint and a
// JWT-guarded group for the current user, URL management and (role-gated)
// user management.
func registerAdmin(
	app *echo.Echo,
	store *urlsvc.URLSvc,
	users *usersvc.UserSvc,
	tokens *auth.TokenService,
	provider *oidc.Service,
	logger *zap.Logger,
	tele telemetry.Telemetery,
) {
	tracer := tele.TraceProvider.Tracer("handler.admin")

	authH := handler.Auth{
		Users:    users,
		Tokens:   tokens,
		Provider: provider,
		Logger:   logger.Named("handler").Named("auth"),
		Tracer:   tracer,
	}
	urlH := handler.AdminURL{
		Store:  store,
		Logger: logger.Named("handler").Named("adminurl"),
		Tracer: tracer,
	}
	userH := handler.AdminUser{
		Users:  users,
		Logger: logger.Named("handler").Named("adminuser"),
		Tracer: tracer,
	}

	api := app.Group("/admin/api")
	api.POST("/auth/login", authH.Login)
	api.GET("/auth/info", authH.AuthInfo)
	api.GET("/auth/oidc/login", authH.OIDCLogin)
	api.GET("/auth/oidc/callback", authH.OIDCCallback)

	authMw := middleware.Auth{Tokens: tokens}
	sec := api.Group("", authMw.Authenticate)

	sec.GET("/auth/me", authH.Me)

	sec.GET("/urls", urlH.List)
	sec.POST("/urls", urlH.Create)
	sec.DELETE("/urls/:key", urlH.Delete)

	sec.GET("/users", userH.List, middleware.RequireRole(model.RoleAdmin))
	sec.POST("/users", userH.Create, middleware.RequireRole(model.RoleSuperAdmin))
	sec.PUT("/users/:id/role", userH.SetRole, middleware.RequireRole(model.RoleSuperAdmin))
	sec.DELETE("/users/:id", userH.Delete, middleware.RequireRole(model.RoleSuperAdmin))
}

// registerSPA serves the embedded admin single-page application under /admin,
// falling back to index.html for client-side routes. Files are written
// directly (rather than via http.FileServer) to avoid its canonical redirects
// for index.html.
func registerSPA(app *echo.Echo) {
	assets := web.Dist()

	serve := func(c *echo.Context) error {
		name := strings.TrimPrefix(c.Request().URL.Path, "/admin")
		name = strings.TrimPrefix(name, "/")

		if name == "" {
			name = "index.html"
		}

		data, err := fs.ReadFile(assets, name)
		if err != nil {
			// Unknown paths fall back to the SPA entrypoint (client-side routing).
			data, err = fs.ReadFile(assets, "index.html")
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "spa not built")
			}

			name = "index.html"
		}

		ctype := mime.TypeByExtension(path.Ext(name))
		if ctype == "" {
			ctype = echo.MIMEOctetStream
		}

		return c.Blob(http.StatusOK, ctype, data)
	}

	app.GET("/admin", serve)
	app.GET("/admin/*", serve)
}
