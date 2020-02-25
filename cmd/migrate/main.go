package migrate

import (
	"context"
	"fmt"

	"github.com/1995parham/koochooloo/config"
	"github.com/1995parham/koochooloo/db"
	"github.com/1995parham/koochooloo/store"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main(cfg config.Config) {
	db, err := db.New(cfg.Database.URL, "urlshortener")
	if err != nil {
		panic(err)
	}

	idx, err := db.Collection(store.Collection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.M{"key": 1},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		panic(err)
	}
	fmt.Println(idx)
}

// Register migrate command
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
