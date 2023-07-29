package urldb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type CommonURLSuite struct {
	suite.Suite

	repo urlrepo.Repository
	app  *fxtest.App
	gen  generator.Generator
}

type MemoryURLSuite struct {
	CommonURLSuite
}

func (suite *MemoryURLSuite) SetupSuite() {
	suite.app = fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(generator.Provide),
		fx.Provide(
			fx.Annotate(urldb.ProvideMemory, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(repo urlrepo.Repository, gen generator.Generator) {
			suite.repo = repo
			suite.gen = gen
		}),
	).RequireStart()
}

func (suite *MemoryURLSuite) TearDownSuite() {
	suite.app.RequireStop()
}

type MongoURLSuite struct {
	CommonURLSuite
}

func (suite *MongoURLSuite) SetupSuite() {
	suite.app = fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(
			fx.Annotate(db.Provide, fx.OnStop(func(ctx context.Context, db *mongo.Database) error {
				_, err := db.Collection(urldb.Collection).DeleteMany(ctx, bson.M{})
				if err != nil {
					return fmt.Errorf("failed to flush records %w", err)
				}

				return nil
			})),
		),
		fx.Provide(generator.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(urldb.ProvideDB, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(repo urlrepo.Repository, gen generator.Generator) {
			suite.repo = repo
			suite.gen = gen
		}),
	).RequireStart()
}

func (suite *MongoURLSuite) TearDownSuite() {
	suite.app.RequireStop()
}

func (suite *CommonURLSuite) TestIncCount() {
	require := suite.Require()

	cases := []struct {
		name   string
		count  int
		inc    int
		expire time.Time
		err    error
	}{
		{
			name:   "Successful",
			count:  2,
			inc:    1,
			expire: time.Time{},
			err:    nil,
		},
		{
			name:   "Expired",
			count:  2,
			inc:    1,
			expire: time.Now().Add(-time.Minute),
			err:    urlrepo.ErrKeyNotFound,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			key := suite.gen.ShortURLKey()

			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			require.NoError(suite.repo.Set(context.Background(), key, "https://elahe-dastan.github.io", expire, c.count))

			for i := 0; i < c.inc; i++ {
				err := suite.repo.Inc(context.Background(), key)
				if c.err == nil {
					require.NoError(err)
				} else {
					require.ErrorIs(err, c.err)
				}
			}

			if c.err == nil {
				count, err := suite.repo.Count(context.Background(), key)
				require.NoError(err)
				require.Equal(c.count+c.inc, count)
			} else {
				_, err := suite.repo.Count(context.Background(), key)
				require.ErrorIs(err, c.err)
			}
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
			key:            "$raha",
			url:            "https://elahe-dastan.github.io",
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Duplicate Key",
			key:            "$raha",
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

			key := c.key
			if key == "" {
				key = suite.gen.ShortURLKey()
			}

			require.ErrorIs(
				suite.repo.Set(context.Background(), key, c.url, expire, 0),
				c.expectedSetErr,
			)

			if c.expectedSetErr == nil {
				url, err := suite.repo.Get(context.Background(), key)
				require.ErrorIs(err, c.expectedGetErr)
				if c.expectedGetErr == nil {
					require.Equal(c.url, url)
				}

				count, err := suite.repo.Count(context.Background(), key)
				require.ErrorIs(err, c.expectedGetErr)
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
