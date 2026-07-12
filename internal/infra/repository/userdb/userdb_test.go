package userdb_test

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CommonUserSuite struct {
	suite.Suite

	repo userrepo.Repository
}

type MemoryUserSuite struct {
	CommonUserSuite
}

func (suite *MemoryUserSuite) SetupTest() {
	suite.repo = userdb.ProvideMemory()
}

type SQLUserSuite struct {
	CommonUserSuite
}

func (suite *SQLUserSuite) SetupTest() {
	//nolint:exhaustruct // only TranslateError is relevant here.
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	suite.Require().NoError(err)

	sqlDB, err := gdb.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)

	suite.Require().NoError(userdb.Migrate(gdb))

	suite.repo = userdb.ProvideDB(gdb, telemetry.ProvideNull(nil))
}

func (suite *CommonUserSuite) TestCreateAndFind() {
	require := suite.Require()
	ctx := suite.T().Context()

	created, err := suite.repo.Create(ctx, model.User{ //nolint:exhaustruct
		Username:     "raha",
		PasswordHash: "hash",
		Role:         model.RoleAdmin,
		Provider:     model.ProviderLocal,
	})
	require.NoError(err)
	require.NotZero(created.ID)

	suite.Run("Duplicate", func() {
		_, err := suite.repo.Create(ctx, model.User{ //nolint:exhaustruct
			Username: "raha",
			Provider: model.ProviderLocal,
		})
		require.ErrorIs(err, userrepo.ErrDuplicateUsername)
	})

	suite.Run("ByUsername", func() {
		u, err := suite.repo.FindByUsername(ctx, "raha")
		require.NoError(err)
		require.Equal(created.ID, u.ID)
		require.Equal(model.RoleAdmin, u.Role)
	})

	suite.Run("ByID", func() {
		u, err := suite.repo.FindByID(ctx, created.ID)
		require.NoError(err)
		require.Equal("raha", u.Username)
	})

	suite.Run("MissingUsername", func() {
		_, err := suite.repo.FindByUsername(ctx, "ghost")
		require.ErrorIs(err, userrepo.ErrUserNotFound)
	})
}

func (suite *CommonUserSuite) TestSubjectLookup() {
	require := suite.Require()
	ctx := suite.T().Context()

	_, err := suite.repo.Create(ctx, model.User{ //nolint:exhaustruct
		Username: "sso-user",
		Role:     model.RoleUser,
		Provider: model.ProviderOIDC,
		Subject:  "sub-123",
	})
	require.NoError(err)

	u, err := suite.repo.FindBySubject(ctx, model.ProviderOIDC, "sub-123")
	require.NoError(err)
	require.Equal("sso-user", u.Username)

	_, err = suite.repo.FindBySubject(ctx, model.ProviderOIDC, "nope")
	require.ErrorIs(err, userrepo.ErrUserNotFound)
}

func (suite *CommonUserSuite) TestRoleAndDelete() {
	require := suite.Require()
	ctx := suite.T().Context()

	created, err := suite.repo.Create(ctx, model.User{ //nolint:exhaustruct
		Username: "changer",
		Role:     model.RoleUser,
		Provider: model.ProviderLocal,
	})
	require.NoError(err)

	require.NoError(suite.repo.SetRole(ctx, created.ID, model.RoleSuperAdmin))

	u, err := suite.repo.FindByID(ctx, created.ID)
	require.NoError(err)
	require.Equal(model.RoleSuperAdmin, u.Role)

	users, err := suite.repo.List(ctx)
	require.NoError(err)
	require.Len(users, 1)

	require.NoError(suite.repo.Delete(ctx, created.ID))

	_, err = suite.repo.FindByID(ctx, created.ID)
	require.ErrorIs(err, userrepo.ErrUserNotFound)

	require.ErrorIs(suite.repo.SetRole(ctx, created.ID, model.RoleAdmin), userrepo.ErrUserNotFound)
	require.ErrorIs(suite.repo.Delete(ctx, created.ID), userrepo.ErrUserNotFound)
}

func TestMemoryUserSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MemoryUserSuite))
}

func TestSQLUserSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SQLUserSuite))
}
