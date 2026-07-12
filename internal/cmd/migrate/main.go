package migrate

import (
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main(logger *zap.Logger, gdb *gorm.DB, shutdonwer fx.Shutdowner) {
	if err := urldb.Migrate(gdb); err != nil {
		panic(err)
	}

	logger.Info("database migrated")

	_ = shutdonwer.Shutdown()
}

// Register migrate command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Setup database schema",
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
