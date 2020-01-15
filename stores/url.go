package stores

import (
	"context"
	"errors"
	"time"

	"github.com/1995parham/koochooloo/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrKeyNotFound indicates that given key does not exist on database
var ErrKeyNotFound = errors.New("given key does not exist or expired")

// ErrDuplicateKey indicates that given key is exists on database
var ErrDuplicateKey = errors.New("given key is exist")

const collection = "urls"

// URLStore communicate with url collections
type URLStore struct {
	DB *mongo.Database
}

// Inc increments counter of url record
func (s URLStore) Inc(ctx context.Context, key string) error {
	record := s.DB.Collection(collection).FindOneAndUpdate(ctx, bson.M{
		"key": key,
	}, bson.M{
		"$inc": bson.M{"count": 1},
	})

	var url models.URL
	if err := record.Decode(&url); err != nil {
		return err
	}
	return nil
}

// Set saves given url with a given key in database
func (s URLStore) Set(ctx context.Context, key string, url string, expire *time.Time) error {
	urls := s.DB.Collection(collection)
	if _, err := urls.InsertOne(ctx, models.URL{
		Key:        key,
		URL:        url,
		ExpireTime: expire,
		Count:      0,
	}); err != nil {
		return err
	}

	return nil
}

// Get retrieves url of the given key if it exists
func (s URLStore) Get(ctx context.Context, key string) (string, error) {
	record := s.DB.Collection(collection).FindOne(ctx, bson.M{
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
	var url models.URL
	if err := record.Decode(&url); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrKeyNotFound
		}
		return "", err
	}

	return url.URL, nil
}
