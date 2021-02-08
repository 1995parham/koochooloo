package logger

import (
	"log"

	"go.uber.org/zap"
)

// New creates a zap logger in debug mode or production.
func New(debug bool) *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)

	if debug {
		logger = zap.NewExample()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		log.Printf("disable logging because of %s\n", err)

		logger = zap.NewNop()
	}

	return logger
}
