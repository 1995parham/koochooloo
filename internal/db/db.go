package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const connectionTimeout = 10 * time.Second

// New creates a new mongodb connection and tests it.
func New(cfg Config) (*mongo.Database, error) {
	// create mongodb connection
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("db new client error: %w", err)
	}

	// connect to the mongodb
	{
		ctx, done := context.WithTimeout(context.Background(), connectionTimeout)
		defer done()

		if err := client.Connect(ctx); err != nil {
			return nil, fmt.Errorf("db connection error: %w", err)
		}
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
