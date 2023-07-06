package seed

import (
	"context"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	store "github.com/1995parham/koochooloo/internal/store/url"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main(cfg *config.Config, logger *zap.Logger) {
	db, err := db.New(cfg.Database)
	if err != nil {
		logger.Fatal("database initiation failed", zap.Error(err))
	}

	str := store.NewMongoURL(db, trace.NewNoopTracerProvider().Tracer(""), noop.NewMeterProvider().Meter(""))

	urls := []string{
		"https://1995parham.me",
		"https://elahe-dastan.github.io",
		"https://github.com/1995parham",
		"https://github.com/elahe-dastan",
	}

	for _, url := range urls {
		if _, err := str.Set(context.Background(), "", url, nil, 0); err != nil {
			logger.Fatal("database insertion failed", zap.Error(err))
		}
	}
}

// Register migrate command.
func Register(root *cobra.Command, cfg *config.Config, logger *zap.Logger) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "seed",
			Short: "Add records into database",
			Run: func(_ *cobra.Command, _ []string) {
				main(cfg, logger)
			},
		},
	)
}
