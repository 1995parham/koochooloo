package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// New creates a new mongodb connection and tests it
func New(url string, db string) (*mongo.Database, error) {
	// create mongodb connection
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("db new client error: %s", err)
	}

	// connect to the mongodb
	ctxc, donec := context.WithTimeout(context.Background(), 10*time.Second)
	defer donec()
	if err := client.Connect(ctxc); err != nil {
		return nil, fmt.Errorf("db connection error: %s", err)
	}

	return client.Database(db), nil
}
