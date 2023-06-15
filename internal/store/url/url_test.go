package url_test

import (
	"context"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/store/url"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type CommonURLSuite struct {
	suite.Suite
	Store url.URL
}

type MemoryURLSuite struct {
	CommonURLSuite
}

func (suite *MemoryURLSuite) SetupSuite() {
	suite.Store = url.NewMemoryURL()
}

func (suite *MemoryURLSuite) TearDownSuite() {
}

type MongoURLSuite struct {
	CommonURLSuite
	DB *mongo.Database
}

func (suite *MongoURLSuite) SetupSuite() {
	cfg := config.New()

	db, err := db.New(cfg.Database)
	suite.Require().NoError(err)

	suite.DB = db
	suite.Store = url.NewMongoURL(db, trace.NewNoopTracerProvider().Tracer(""), noop.NewMeterProvider().Meter(""))
}

func (suite *MongoURLSuite) TearDownSuite() {
	_, err := suite.DB.Collection(url.Collection).DeleteMany(context.Background(), bson.D{})
	suite.Require().NoError(err)

	suite.Require().NoError(suite.DB.Client().Disconnect(context.Background()))
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
			key, err := suite.Store.Set(context.Background(), "", "https://elahe-dastan.github.io", nil, c.count)
			require.NoError(err)

			for i := 0; i < c.inc; i++ {
				require.NoError(suite.Store.Inc(context.Background(), key))
			}

			count, err := suite.Store.Count(context.Background(), key)
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
			expectedSetErr: url.ErrDuplicateKey,
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
			expectedGetErr: url.ErrKeyNotFound,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			key, err := suite.Store.Set(context.Background(), c.key, c.url, expire, 0)
			require.Equal(c.expectedSetErr, err)

			if c.expectedSetErr == nil {
				if c.key != "" {
					require.Equal("$"+c.key, key)
				}

				url, err := suite.Store.Get(context.Background(), key)
				require.Equal(c.expectedGetErr, err)
				if c.expectedGetErr == nil {
					require.Equal(c.url, url)
				}

				count, err := suite.Store.Count(context.Background(), key)
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
