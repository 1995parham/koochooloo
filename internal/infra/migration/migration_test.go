package migration_test

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/migration"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func open(t *testing.T) *gorm.DB {
	t.Helper()

	//nolint:exhaustruct_v5 // only TranslateError matters for these tests.
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	require.NoError(t, err)

	return gdb
}

func TestOnStartDisabled(t *testing.T) {
	t.Parallel()

	gdb := open(t)

	cfg := db.Config{Dialect: "sqlite", URL: ":memory:", MigrateOnStart: false}

	require.NoError(t, migration.OnStart(cfg, gdb, zap.NewNop()))

	tables, err := gdb.Migrator().GetTables()
	require.NoError(t, err)
	require.Empty(t, tables)
}

func TestOnStartEnabledIsIdempotent(t *testing.T) {
	t.Parallel()

	gdb := open(t)
	cfg := db.Config{Dialect: "sqlite", URL: ":memory:", MigrateOnStart: true}

	require.NoError(t, migration.OnStart(cfg, gdb, zap.NewNop()))

	require.True(t, gdb.Migrator().HasTable("users"))
	require.True(t, gdb.Migrator().HasTable("urls"))

	// A restart runs it again against the migrated schema.
	require.NoError(t, migration.OnStart(cfg, gdb, zap.NewNop()))
}
