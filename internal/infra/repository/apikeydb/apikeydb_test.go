package apikeydb_test

import (
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/apikeyrepo"
	"github.com/1995parham/koochooloo/internal/infra/repository/apikeydb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CommonAPIKeySuite struct {
	suite.Suite

	repo apikeyrepo.Repository
}

type MemoryAPIKeySuite struct {
	CommonAPIKeySuite
}

func (suite *MemoryAPIKeySuite) SetupTest() {
	suite.repo = apikeydb.ProvideMemory()
}

type SQLAPIKeySuite struct {
	CommonAPIKeySuite
}

func (suite *SQLAPIKeySuite) SetupTest() {
	//nolint:exhaustruct_v5 // only TranslateError is relevant here.
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	suite.Require().NoError(err)

	sqlDB, err := gdb.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)

	suite.Require().NoError(apikeydb.Migrate(gdb))

	suite.repo = apikeydb.ProvideDB(gdb, telemetry.ProvideNull(nil))
}

func (suite *CommonAPIKeySuite) TestLifecycle() {
	require := suite.Require()
	ctx := suite.T().Context()

	created, err := suite.repo.Create(ctx, model.APIKey{
		ID:         0,
		UserID:     3,
		Name:       "hatchup-team",
		Prefix:     "kch_abcdef",
		Hash:       "deadbeef",
		CreatedAt:  time.Now(),
		LastUsedAt: nil,
		RevokedAt:  nil,
	})
	require.NoError(err)
	require.NotZero(created.ID)

	found, err := suite.repo.FindByHash(ctx, "deadbeef")
	require.NoError(err)
	require.Equal(created.ID, found.ID)
	require.Equal(uint(3), found.UserID)
	require.False(found.Revoked())

	_, err = suite.repo.FindByHash(ctx, "missing")
	require.ErrorIs(err, apikeyrepo.ErrAPIKeyNotFound)

	used := time.Now().UTC().Truncate(time.Second)
	require.NoError(suite.repo.Touch(ctx, created.ID, used))
	require.NoError(suite.repo.Revoke(ctx, created.ID, used))
	require.ErrorIs(suite.repo.Revoke(ctx, 999, used), apikeyrepo.ErrAPIKeyNotFound)

	keys, err := suite.repo.List(ctx)
	require.NoError(err)
	require.Len(keys, 1)
	require.NotNil(keys[0].LastUsedAt)
	require.True(keys[0].Revoked())
}

func TestMemoryAPIKeySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MemoryAPIKeySuite))
}

func TestSQLAPIKeySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SQLAPIKeySuite))
}
