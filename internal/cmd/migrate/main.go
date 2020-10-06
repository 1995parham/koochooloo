package migrate

import (
	"context"
	"fmt"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/store"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const enable = 1

func main(cfg config.Config) {
	db, err := db.New(cfg.Database.URL, cfg.Database.Name)
	if err != nil {
		panic(err)
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

	fmt.Println(idx)
}

// Register migrate command.
func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "migrate",
			Short: "Setup database indices",
			Run: func(cmd *cobra.Command, args []string) {
				main(cfg)
			},
		},
	)
}
