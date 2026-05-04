package urldb_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	domgen "github.com/1995parham/koochooloo/internal/domain/generator"
	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const testURL = "https://elahe-dastan.github.io"

type CommonURLSuite struct {
	suite.Suite

	repo urlrepo.Repository
	app  *fxtest.App
	gen  domgen.Generator
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
		fx.Invoke(func(repo urlrepo.Repository, gen domgen.Generator) {
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
		fx.Invoke(func(repo urlrepo.Repository, gen domgen.Generator) {
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
	context := suite.T().Context()

	cases := []struct {
		name   string
		count  int // current number of visits
		inc    int // expected increase in the number of visits
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
			inc:    0,
			expire: time.Now().Add(-time.Minute),
			err:    urlrepo.ErrKeyNotFound,
		},
	}

	for _, c := range cases {
		suite.Run(c.name, func() {
			key := suite.gen.ShortURLKey()

			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			require.NoError(suite.repo.Save(context, model.URL{
				Key:        key,
				URL:        testURL,
				ExpireTime: expire,
				Count:      c.count,
			}))

			for range c.inc {
				err := suite.repo.IncrementCount(context, key)
				if c.err == nil {
					require.NoError(err)
				} else {
					require.ErrorIs(err, c.err)
				}
			}

			if c.err == nil {
				u, err := suite.repo.FindByKey(context, key)
				require.NoError(err)
				require.Equal(c.count+c.inc, u.Count)
			} else {
				_, err := suite.repo.FindByKey(context, key)
				require.ErrorIs(err, c.err)
			}
		})
	}
}

// nolint: funlen
func (suite *CommonURLSuite) TestIncrementConsistency() {
	require := suite.Require()
	ctx := suite.T().Context()

	suite.Run("MultipleIncrements", func() {
		key := suite.gen.ShortURLKey()

		require.NoError(suite.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        "https://example.com",
			Count:      0,
			ExpireTime: nil,
		}))

		const increments = 50

		for range increments {
			require.NoError(suite.repo.IncrementCount(ctx, key))
		}

		u, err := suite.repo.FindByKey(ctx, key)
		require.NoError(err)
		require.Equal(increments, u.Count)
	})

	suite.Run("IncrementPreservesFields", func() {
		key := suite.gen.ShortURLKey()
		originalURL := "https://preserve-me.com"

		require.NoError(suite.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        originalURL,
			Count:      5,
			ExpireTime: nil,
		}))

		require.NoError(suite.repo.IncrementCount(ctx, key))

		u, err := suite.repo.FindByKey(ctx, key)
		require.NoError(err)
		require.Equal(6, u.Count)
		require.Equal(originalURL, u.URL)
		require.Equal(key, u.Key)
	})

	suite.Run("IncrementNonExistentKey", func() {
		err := suite.repo.IncrementCount(ctx, "does-not-exist")
		require.ErrorIs(err, urlrepo.ErrKeyNotFound)
	})

	suite.Run("IncrementFromInitialCount", func() {
		key := suite.gen.ShortURLKey()

		require.NoError(suite.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        "https://initial-count.com",
			Count:      100,
			ExpireTime: nil,
		}))

		require.NoError(suite.repo.IncrementCount(ctx, key))
		require.NoError(suite.repo.IncrementCount(ctx, key))
		require.NoError(suite.repo.IncrementCount(ctx, key))

		u, err := suite.repo.FindByKey(ctx, key)
		require.NoError(err)
		require.Equal(103, u.Count)
	})

	suite.Run("ConcurrentIncrements", func() {
		key := suite.gen.ShortURLKey()

		require.NoError(suite.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        "https://concurrent.com",
			Count:      0,
			ExpireTime: nil,
		}))

		const (
			goroutines = 10
			perWorker  = 20
		)

		var wg sync.WaitGroup

		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()

				for range perWorker {
					_ = suite.repo.IncrementCount(ctx, key)
				}
			}()
		}

		wg.Wait()

		u, err := suite.repo.FindByKey(ctx, key)
		require.NoError(err)
		require.Equal(goroutines*perWorker, u.Count)
	})
}

// nolint: funlen
func (suite *CommonURLSuite) TestSetGetCount() {
	require := suite.Require()
	context := suite.T().Context()

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
			url:            testURL,
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Duplicate Key",
			key:            "$raha",
			url:            testURL,
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
		suite.Run(c.name, func() {
			expire := new(time.Time)

			*expire = c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			key := c.key
			if key == "" {
				key = suite.gen.ShortURLKey()
			}

			require.ErrorIs(
				suite.repo.Save(context, model.URL{
					Key:        key,
					URL:        c.url,
					ExpireTime: expire,
					Count:      0,
				}),
				c.expectedSetErr,
			)

			if c.expectedSetErr == nil {
				u, err := suite.repo.FindByKey(context, key)
				require.ErrorIs(err, c.expectedGetErr)

				if c.expectedGetErr == nil {
					require.Equal(c.url, u.URL)
					require.Equal(0, u.Count)
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
