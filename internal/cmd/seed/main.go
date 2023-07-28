package seed

import (
	"context"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main(logger *zap.Logger, repo urlrepo.Repository, shutdowner fx.Shutdowner) {
	urls := []string{
		"https://1995parham.me",
		"https://elahe-dastan.github.io",
		"https://github.com/1995parham",
		"https://github.com/elahe-dastan",
	}

	for _, url := range urls {
		if _, err := repo.Set(context.Background(), "", url, nil, 0); err != nil {
			logger.Fatal("database insertion failed", zap.Error(err))
		}
	}

	_ = shutdowner.Shutdown()
}

// Register migrate command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "seed",
			Short: "Add records into database",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.Provide(db.Provide),
					fx.Provide(telemetry.ProvideNull),
					fx.Provide(
						fx.Annotate(urldb.ProvideDB, fx.As(new(urlrepo.Repository))),
					),
					fx.NopLogger,
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
