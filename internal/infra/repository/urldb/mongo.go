package urldb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.opentelemetry.io/otel/trace"
)

// MongoURL communicate with url collections in MongoDB.
type MongoURL struct {
	DB      *mongo.Database
	Tracer  trace.Tracer
	Metrics Metrics
}

// Collection is a name of the MongoDB collection for URLs.
const (
	Collection = "urls"
	one        = 1
)

// ProvideDB creates new URL mongodb store.
func ProvideDB(db *mongo.Database, tele telemetry.Telemetery) *MongoURL {
	tracer := tele.TraceProvider.Tracer("urldb.db")
	meter := tele.MeterProvider.Meter("urldb.db")

	return &MongoURL{
		DB:     db,
		Tracer: tracer,
		Metrics: Metrics{
			Usage:   NewUsage(meter, "mongo"),
			Latency: NewLatency(meter),
		},
	}
}

// liveFilter returns a BSON filter that matches a key only if it is not expired.
func liveFilter(key string) bson.M {
	return bson.M{
		"key": key,
		"$or": bson.A{
			bson.M{"expire_time": bson.M{"$eq": nil}},
			bson.M{"expire_time": bson.M{"$gte": time.Now()}},
		},
	}
}

// IncrementCount increments counter of url record by one, means url got visited.
func (s *MongoURL) IncrementCount(ctx context.Context, key string) error {
	ctx, span := s.Tracer.Start(ctx, "store.url.inc")
	defer span.End()

	record := s.DB.Collection(Collection).FindOneAndUpdate(ctx,
		liveFilter(key),
		bson.M{"$inc": bson.M{"count": one}},
	)

	var doc urlDocument
	if err := record.Decode(&doc); err != nil {
		span.RecordError(err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return urlrepo.ErrKeyNotFound
		}

		return fmt.Errorf("mongodb failed: %w", err)
	}

	return nil
}

// Save saves given url in database.
func (s *MongoURL) Save(ctx context.Context, url model.URL) error {
	ctx, span := s.Tracer.Start(ctx, "store.url.save")
	defer span.End()

	urls := s.DB.Collection(Collection)

	start := time.Now()

	if _, err := urls.InsertOne(ctx, toDocument(url)); err != nil {
		span.RecordError(err)

		if mongo.IsDuplicateKeyError(err) {
			return urlrepo.ErrDuplicateKey
		}

		return fmt.Errorf("mongodb failed: %w", err)
	}

	s.Metrics.InsertedCounter.Add(ctx, 1)
	s.Metrics.InsertLatency.Record(ctx, time.Since(start).Seconds())

	return nil
}

// FindByKey retrieves url of the given key if it exists.
func (s *MongoURL) FindByKey(ctx context.Context, key string) (model.URL, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.find_by_key")
	defer span.End()

	record := s.DB.Collection(Collection).FindOne(ctx, liveFilter(key))

	var doc urlDocument
	if err := record.Decode(&doc); err != nil {
		span.RecordError(err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.URL{}, urlrepo.ErrKeyNotFound
		}

		return model.URL{}, fmt.Errorf("mongodb failed: %w", err)
	}

	s.Metrics.FetchedCounter.Add(ctx, 1)

	return toModel(doc), nil
}
