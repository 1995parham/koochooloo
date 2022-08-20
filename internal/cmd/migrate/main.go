package migrate

import (
	"context"

	"github.com/1995parham/koochooloo/internal"
	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	store "github.com/1995parham/koochooloo/internal/store/url"
	"github.com/1995parham/koochooloo/pkg/telemetry"
	"github.com/1995parham/koochooloo/pkg/telemetry/log"
	"github.com/1995parham/koochooloo/pkg/telemetry/metric"
	"github.com/1995parham/koochooloo/pkg/telemetry/trace"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const enable = 1

func main(cfg *config.Config) {
	telemetry := &telemetry.Telemetry{
		Log:    log.NewZap(cfg.Telemetry.Log),
		Metric: metric.New(internal.Namespace, internal.Subsystem),
		Trace:  trace.New(cfg.Telemetry.Trace, internal.Namespace, internal.Subsystem),
	}

	db, err := db.New(cfg.Database)
	if err != nil {
		telemetry.Log.Fatal("database initiation failed", zap.Error(err))
	}

	idx, err := db.Collection(store.Collection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"key": enable},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		panic(err)
	}

	telemetry.Log.Info("database index", zap.Any("index", idx))
}

// Register migrate command.
func Register(root *cobra.Command, cfg *config.Config) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Setup database indices",
			Run:   func(cmd *cobra.Command, args []string) { main(cfg) },
		},
	)
}
