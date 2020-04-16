package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrKeyNotFound indicates that given key does not exist on database
var ErrKeyNotFound = errors.New("given key does not exist or expired")

// ErrDuplicateKey indicates that given key is exists on database
var ErrDuplicateKey = errors.New("given key is exist")

// URL stores and retrieves urls
type URL interface {
	Inc(ctx context.Context, key string) error
	Set(ctx context.Context, key string, url string, expire *time.Time) (string, error)
	Get(ctx context.Context, key string) (string, error)
}

// Collection is a name of the MongoDB collection for URLs
const Collection = "urls"
const one = 1

// MongoURL communicate with url collections in MongoDB
type MongoURL struct {
	DB *mongo.Database
	Usage
}

// NewMongoURL creates new URL store
func NewMongoURL(db *mongo.Database) *MongoURL {
	return &MongoURL{
		DB:    db,
		Usage: NewUsage("url"),
	}
}

// Inc increments counter of url record by one
func (s *MongoURL) Inc(ctx context.Context, key string) error {
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
func (s *MongoURL) Set(ctx context.Context, key string, url string, expire *time.Time) (string, error) {
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
		if !strings.HasPrefix(key, "$") && errors.Is(err, ErrDuplicateKey) {
			return s.Set(ctx, "", url, expire)
		}

		return "", err
	}

	return key, nil
}

// Get retrieves url of the given key if it exists
func (s *MongoURL) Get(ctx context.Context, key string) (string, error) {
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
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrKeyNotFound
		}

		return "", err
	}

	s.FetchedCounter.Inc()

	return url.URL, nil
}
