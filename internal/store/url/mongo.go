package url

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

// MongoURL communicate with url collections in MongoDB.
type MongoURL struct {
	DB      *mongo.Database
	Tracer  trace.Tracer
	Metrics Usage
}

// Collection is a name of the MongoDB collection for URLs.
const (
	Collection = "urls"
	one        = 1
)

// NewMongoURL creates new URL store.
func NewMongoURL(db *mongo.Database, tracer trace.Tracer) *MongoURL {
	return &MongoURL{
		DB:      db,
		Tracer:  tracer,
		Metrics: NewUsage("mongo"),
	}
}

// Inc increments counter of url record by one.
func (s *MongoURL) Inc(ctx context.Context, key string) error {
	ctx, span := s.Tracer.Start(ctx, "store.url.inc")
	defer span.End()

	record := s.DB.Collection(Collection).FindOneAndUpdate(ctx, bson.M{
		"key": key,
	}, bson.M{
		"$inc": bson.M{"count": one},
	})

	var url model.URL
	if err := record.Decode(&url); err != nil {
		span.RecordError(err)

		return fmt.Errorf("mongodb failed: %w", err)
	}

	return nil
}

// Set saves given url with a given key in database. if key is null it generates a random key and returns it.
func (s *MongoURL) Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.set")
	defer span.End()

	if key == "" {
		key = model.Key()
	} else {
		key = fmt.Sprintf("$%s", key)
	}

	urls := s.DB.Collection(Collection)

	if _, err := urls.InsertOne(ctx, model.URL{
		Key:        key,
		URL:        url,
		ExpireTime: expire,
		Count:      count,
	}); err != nil {
		span.RecordError(err)

		if mongo.IsDuplicateKeyError(err) {
			if !strings.HasPrefix(key, "$") {
				// call set again to generate another random key.
				return s.Set(ctx, "", url, expire, 0)
			}

			return "", ErrDuplicateKey
		}

		return "", fmt.Errorf("mongodb failed: %w", err)
	}

	s.Metrics.InsertedCounter.Inc()

	return key, nil
}

// Get retrieves url of the given key if it exists.
func (s *MongoURL) Get(ctx context.Context, key string) (string, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.get")
	defer span.End()

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
		span.RecordError(err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrKeyNotFound
		}

		return "", fmt.Errorf("mongodb failed: %w", err)
	}

	s.Metrics.FetchedCounter.Inc()

	return url.URL, nil
}

// Count retrieves number of access for the url of the given key if it exists.
func (s *MongoURL) Count(ctx context.Context, key string) (int, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.count")
	defer span.End()

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
		span.RecordError(err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, ErrKeyNotFound
		}

		return 0, fmt.Errorf("mongodb failed: %w", err)
	}

	return count.Count, nil
}
