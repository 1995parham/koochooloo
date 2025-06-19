package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
	"go.uber.org/fx"
)

const connectionTimeout = 10 * time.Second

// New creates a new mongodb connection and tests it.
func Provide(lc fx.Lifecycle, cfg Config) (*mongo.Database, error) {
	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor()
	opts.ApplyURI(cfg.URL)
	opts.SetConnectTimeout(connectionTimeout)

	// connect to the mongodb
	client, err := mongo.Connect(opts)
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

	lc.Append(
		fx.StopHook(client.Disconnect),
	)

	return client.Database(cfg.Name), nil
}
