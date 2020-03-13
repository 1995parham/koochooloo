package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrKeyNotFound indicates that given key does not exist on database
var ErrKeyNotFound = errors.New("given key does not exist or expired")

// ErrDuplicateKey indicates that given key is exists on database
var ErrDuplicateKey = errors.New("given key is exist")

// Collection is a name of the MongoDB collection for URLs
const Collection = "urls"
const one = 1

// URL communicate with url collections in MongoDB
type URL struct {
	DB *mongo.Database

	InsertedCounter prometheus.Counter
	FetchedCounter  prometheus.Counter
}

// NewURL creates new URL store
func NewURL(db *mongo.Database) *URL {
	return &URL{
		DB: db,
		InsertedCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "koochooloo",
			Name:      "inserted_urls_counter",
		}),
		FetchedCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "koochooloo",
			Name:      "fetched_urls_counter",
		}),
	}
}

// Inc increments counter of url record by one
func (s *URL) Inc(ctx context.Context, key string) error {
	record := s.DB.Collection(Collection).FindOneAndUpdate(ctx, bson.M{
		"key": key,
	}, bson.M{
		"$inc": bson.M{"count": one},
	})

	var url model.URL
	if err := record.Decode(&url); err != nil {
		return err
	}

	return nil
}

// Set saves given url with a given key in database. if key is null it generates a random key and returns it.
func (s *URL) Set(ctx context.Context, key string, url string, expire *time.Time) (string, error) {
	if key == "" {
		key = Key()
	} else {
		key = fmt.Sprintf("$%s", key)
	}

	s.InsertedCounter.Inc()

	urls := s.DB.Collection(Collection)

	_, err := urls.InsertOne(ctx, model.URL{
		Key:        key,
		URL:        url,
		ExpireTime: expire,
		Count:      0,
	})
	if err != nil {
		if !strings.HasPrefix(key, "$") && err == ErrDuplicateKey {
			return s.Set(ctx, "", url, expire)
		}

		return "", err
	}

	return key, nil
}

// Get retrieves url of the given key if it exists
func (s *URL) Get(ctx context.Context, key string) (string, error) {
	record := s.DB.Collection(Collection).FindOne(ctx, bson.M{
		"key": key,
		"$or": bson.A{
			bson.M{
				"expire_time": bson.M{
					"$eq": nil,
				},
			},
			bson.M{
				"expire_time": bson.M{
					"$gte": time.Now(),
				},
			},
		},
	})

	var url model.URL
	if err := record.Decode(&url); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrKeyNotFound
		}

		return "", err
	}

	s.FetchedCounter.Inc()

	return url.URL, nil
}
