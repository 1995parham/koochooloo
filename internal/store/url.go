package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrKeyNotFound indicates that given key does not exist on database.
	ErrKeyNotFound = errors.New("given key does not exist or expired")
	// ErrDuplicateKey indicates that given key is exists on database.
	ErrDuplicateKey = errors.New("given key is exist")
)

type (
	// URL stores and retrieves urls.
	URL interface {
		Inc(ctx context.Context, key string) error
		Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error)
		Get(ctx context.Context, key string) (string, error)
		Count(ctx context.Context, key string) (int, error)
	}
	// MongoURL communicate with url collections in MongoDB.
	MongoURL struct {
		DB     *mongo.Database
		Tracer trace.Tracer
		Usage
	}
)

// Collection is a name of the MongoDB collection for URLs.
const (
	Collection                   = "urls"
	one                          = 1
	mongodbDuplicateKeyErrorCode = 11000
)

// NewMongoURL creates new URL store.
func NewMongoURL(db *mongo.Database, tracer trace.Tracer) *MongoURL {
	return &MongoURL{
		DB:     db,
		Tracer: tracer,
		Usage:  NewUsage("url"),
	}
}

// Inc increments counter of url record by one.
func (s *MongoURL) Inc(ctx context.Context, key string) error {
	record := s.DB.Collection(Collection).FindOneAndUpdate(ctx, bson.M{
		"key": key,
	}, bson.M{
		"$inc": bson.M{"count": one},
	})

	var url model.URL
	if err := record.Decode(&url); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// Set saves given url with a given key in database. if key is null it generates a random key and returns it.
func (s *MongoURL) Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.set")
	defer span.End()

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
		Count:      count,
	})
	if err != nil {
		var exp mongo.WriteException

		if ok := errors.As(err, &exp); ok &&
			exp.WriteErrors[0].Code == mongodbDuplicateKeyErrorCode {
			if !strings.HasPrefix(key, "$") {
				return s.Set(ctx, "", url, expire, 0)
			}

			return "", ErrDuplicateKey
		}

		return "", fmt.Errorf("%w", err)
	}

	return key, nil
}

// Get retrieves url of the given key if it exists.
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

		return "", fmt.Errorf("%w", err)
	}

	s.FetchedCounter.Inc()

	return url.URL, nil
}

// Count retrieves number of access for the url of the given key if it exists.
func (s *MongoURL) Count(ctx context.Context, key string) (int, error) {
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
	}, options.FindOne().SetProjection(bson.M{"count": true}))

	var count struct {
		Count int
	}

	if err := record.Decode(&count); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, ErrKeyNotFound
		}

		return 0, fmt.Errorf("%w", err)
	}

	return count.Count, nil
}
