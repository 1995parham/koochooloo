package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/store"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoURLSuite struct {
	suite.Suite
	DB    *mongo.Database
	Store store.URL
}

func (suite *MongoURLSuite) SetupSuite() {
	cfg := config.New()

	db, err := db.New(cfg.Database.URL, cfg.Database.Name)
	suite.NoError(err)

	suite.DB = db
	suite.Store = store.NewMongoURL(db)
}

func (suite *MongoURLSuite) TearDownSuite() {
	_, err := suite.DB.Collection(store.Collection).DeleteMany(context.Background(), bson.D{})
	suite.NoError(err)

	suite.NoError(suite.DB.Client().Disconnect(context.Background()))
}

func (suite *MongoURLSuite) TestIncCount() {
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
			suite.NoError(err)

			for i := 0; i < c.inc; i++ {
				suite.NoError(suite.Store.Inc(context.Background(), key))
			}

			count, err := suite.Store.Count(context.Background(), key)
			suite.NoError(err)
			suite.Equal(c.count+c.inc, count)
		})
	}
}

// nolint: funlen
func (suite *MongoURLSuite) TestSetGetCount() {
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
			expectedSetErr: store.ErrDuplicateKey,
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
			expectedGetErr: store.ErrKeyNotFound,
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
			suite.Equal(c.expectedSetErr, err)

			if c.expectedSetErr == nil {
				if c.key != "" {
					suite.Equal("$"+c.key, key)
				}

				url, err := suite.Store.Get(context.Background(), key)
				suite.Equal(c.expectedGetErr, err)
				if c.expectedGetErr == nil {
					suite.Equal(c.url, url)
				}

				count, err := suite.Store.Count(context.Background(), key)
				suite.Equal(c.expectedGetErr, err)
				if c.expectedGetErr == nil {
					suite.Equal(0, count)
				}
			}
		})
	}
}

func TestMongoURLSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MongoURLSuite))
}
