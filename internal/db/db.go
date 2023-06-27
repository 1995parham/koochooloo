package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const connectionTimeout = 10 * time.Second

// New creates a new mongodb connection and tests it.
func New(cfg Config) (*mongo.Database, error) {
	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor()
	opts.ApplyURI(cfg.URL)

	// connect to the mongodb
	ctx, done := context.WithTimeout(context.Background(), connectionTimeout)
	defer done()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	// ping the mongodb
	{
		ctx, done := context.WithTimeout(context.Background(), connectionTimeout)
		defer done()

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			return nil, fmt.Errorf("db ping error: %w", err)
		}
	}

	return client.Database(cfg.Name), nil
}
