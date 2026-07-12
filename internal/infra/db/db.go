package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const connectionTimeout = 10 * time.Second

// ErrUnsupportedDialect indicates that the configured database dialect is not
// one of the engines koochooloo knows how to open.
var ErrUnsupportedDialect = errors.New("unsupported database dialect")

// dialector returns the GORM dialector for the configured engine. Adding a new
// engine is a matter of importing its driver and adding a case here.
func dialector(cfg Config) (gorm.Dialector, error) {
	switch cfg.Dialect {
	case "sqlite":
		return sqlite.Open(cfg.URL), nil
	case "postgres":
		return postgres.Open(cfg.URL), nil
	case "mysql":
		return mysql.Open(cfg.URL), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedDialect, cfg.Dialect)
	}
}

// Provide opens a GORM connection for the configured dialect, verifies it with
// a ping, and registers a stop hook to close the underlying pool. The same
// models and queries run against any supported engine — only cfg.Dialect and
// cfg.URL change.
func Provide(lc fx.Lifecycle, cfg Config) (*gorm.DB, error) {
	dial, err := dialector(cfg)
	if err != nil {
		return nil, err
	}

	//nolint:exhaustruct // gorm.Config is opt-in; only the fields we set matter.
	gdb, err := gorm.Open(dial, &gorm.Config{
		// TranslateError maps engine-specific driver errors onto GORM sentinels
		// (gorm.ErrDuplicatedKey, gorm.ErrRecordNotFound, ...) so the repository
		// stays dialect-agnostic.
		TranslateError: true,
		Logger:         gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("db handle error: %w", err)
	}

	// ping the database
	{
		ctx, done := context.WithTimeout(context.Background(), connectionTimeout)
		defer done()

		if err := sqlDB.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("db ping error: %w", err)
		}
	}

	lc.Append(
		fx.StopHook(sqlDB.Close),
	)

	return gdb, nil
}
