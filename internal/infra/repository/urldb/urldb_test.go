package urldb_test

import (
	"context"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type CommonURLSuite struct {
	suite.Suite

	repo urlrepo.Repository
}

type MemoryURLSuite struct {
	CommonURLSuite
}

func (suite *MemoryURLSuite) SetupSuite() {
	fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(urldb.ProvideMemory, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(repo urlrepo.Repository) {
			suite.repo = repo
		}),
	).RequireStart().RequireStop()
}

func (suite *MemoryURLSuite) TearDownSuite() {}

type MongoURLSuite struct {
	CommonURLSuite
}

func (suite *MongoURLSuite) SetupSuite() {
	fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(urldb.ProvideDB, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(repo urlrepo.Repository) {
			suite.repo = repo
		}),
	).RequireStart().RequireStop()
}

func (suite *CommonURLSuite) TestIncCount() {
	require := suite.Require()

	cases := []struct {
		name  string
		count int
		inc   int
	}{
		{
			name:  "Successful",
			count: 2,
			inc:   1,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			key, err := suite.repo.Set(context.Background(), "", "https://elahe-dastan.github.io", nil, c.count)
			require.NoError(err)

			for i := 0; i < c.inc; i++ {
				require.NoError(suite.repo.Inc(context.Background(), key))
			}

			count, err := suite.repo.Count(context.Background(), key)
			require.NoError(err)
			require.Equal(c.count+c.inc, count)
		})
	}
}

// nolint: funlen
func (suite *CommonURLSuite) TestSetGetCount() {
	require := suite.Require()

	cases := []struct {
		name           string
		key            string
		url            string
		expire         time.Time
		expectedSetErr error
		expectedGetErr error
	}{
		{
			name:           "Successful",
			key:            "raha",
			url:            "https://elahe-dastan.github.io",
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Duplicate Key",
			key:            "raha",
			url:            "https://elahe-dastan.github.io",
			expire:         time.Time{},
			expectedSetErr: urlrepo.ErrDuplicateKey,
			expectedGetErr: nil,
		},
		{
			name:           "Automatic",
			key:            "",
			url:            "https://1995parham.me",
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Expired",
			key:            "",
			url:            "https://github.com",
			expire:         time.Now().Add(-time.Minute),
			expectedSetErr: nil,
			expectedGetErr: urlrepo.ErrKeyNotFound,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			key, err := suite.repo.Set(context.Background(), c.key, c.url, expire, 0)
			require.Equal(c.expectedSetErr, err)

			if c.expectedSetErr == nil {
				if c.key != "" {
					require.Equal("$"+c.key, key)
				}

				url, err := suite.repo.Get(context.Background(), key)
				require.Equal(c.expectedGetErr, err)
				if c.expectedGetErr == nil {
					require.Equal(c.url, url)
				}

				count, err := suite.repo.Count(context.Background(), key)
				require.Equal(c.expectedGetErr, err)
				if c.expectedGetErr == nil {
					require.Equal(0, count)
				}
			}
		})
	}
}

func TestMongoURLSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MongoURLSuite))
}

func TestMemoryURLSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MemoryURLSuite))
}
