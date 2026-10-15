// Package migration owns the database schema setup shared by the `migrate`
// command and the server's opt-in migrate-on-start.
package migration

import (
	"fmt"

	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/repository/apikeydb"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Run creates or updates every table koochooloo needs. GORM's AutoMigrate
// only adds what is missing, so running it on a migrated database is a no-op.
func Run(gdb *gorm.DB) error {
	if err := userdb.Migrate(gdb); err != nil {
		return fmt.Errorf("migrating users: %w", err)
	}

	if err := urldb.Migrate(gdb); err != nil {
		return fmt.Errorf("migrating urls: %w", err)
	}

	if err := apikeydb.Migrate(gdb); err != nil {
		return fmt.Errorf("migrating api keys: %w", err)
	}

	return nil
}

// OnStart runs the migration when cfg.MigrateOnStart is set. The server
// invokes it while the fx graph is built, so it finishes before the HTTP
// server's start hook begins listening — the hook a distroless image without
// a shell has for setting up its schema.
func OnStart(cfg db.Config, gdb *gorm.DB, logger *zap.Logger) error {
	if !cfg.MigrateOnStart {
		return nil
	}

	if err := Run(gdb); err != nil {
		return err
	}

	logger.Info("database migrated on start")

	return nil
}
