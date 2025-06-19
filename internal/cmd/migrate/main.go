package migrate

import (
	"context"

	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const enable = 1

func main(logger *zap.Logger, db *mongo.Database, shutdonwer fx.Shutdowner) {
	idx, err := db.Collection(urldb.Collection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"key": enable},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		panic(err)
	}

	logger.Info("database index", zap.Any("index", idx))

	_ = shutdonwer.Shutdown()
}

// Register migrate command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Setup database indices",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(db.Provide),
					fx.NopLogger,
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
