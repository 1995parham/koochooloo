// Package apikey provides the `koochooloo apikey` command group for minting,
// listing and revoking the API keys services use to call the admin API.
package apikey

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/apikeyrepo"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/apikeysvc"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/apikeydb"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register adds the apikey command group to the root command.
func Register(root *cobra.Command) {
	//nolint:exhaustruct_v5
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "Manage API keys for services that call the admin API",
	}

	cmd.AddCommand(createCommand(), listCommand(), revokeCommand())
	root.AddCommand(cmd)
}

// runFx builds the API-key service fx graph and invokes fn against it.
func runFx(invoke any) {
	fx.New(
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(db.Provide),
		fx.Provide(fx.Annotate(userdb.ProvideDB, fx.As(new(userrepo.Repository)))),
		fx.Provide(fx.Annotate(apikeydb.ProvideDB, fx.As(new(apikeyrepo.Repository)))),
		fx.Provide(apikeysvc.Provide),
		fx.NopLogger,
		fx.Invoke(invoke),
	).Run()
}

func createCommand() *cobra.Command {
	var username, name string

	//nolint:exhaustruct_v5
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API key for a user and print it once",
		Long: "Create an API key that acts as the given user. The key is printed on stdout " +
			"exactly once and only its hash is stored, so save it right away.",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *apikeysvc.APIKeySvc, sd fx.Shutdowner, logger *zap.Logger) {
				plain, key, err := svc.Create(context.Background(), username, name)
				if err != nil {
					logger.Error("failed to create api key", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				logger.Info("api key created",
					zap.Uint("id", key.ID),
					zap.String("name", key.Name),
					zap.String("prefix", key.Prefix),
				)

				// Only the key goes to stdout, so `$(koochooloo apikey create ...)` captures it.
				_, _ = fmt.Fprintln(os.Stdout, plain)

				_ = sd.Shutdown()
			})
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "user the key acts as (required)")
	cmd.Flags().StringVar(&name, "name", "", "what the key is for, e.g. hatchup-team (required)")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func listCommand() *cobra.Command {
	//nolint:exhaustruct_v5
	return &cobra.Command{
		Use:   "list",
		Short: "List API keys (never the keys themselves)",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *apikeysvc.APIKeySvc, sd fx.Shutdowner, logger *zap.Logger) {
				keys, err := svc.List(context.Background())
				if err != nil {
					logger.Error("failed to list api keys", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				_, _ = fmt.Fprintf(os.Stdout, "%-5s %-8s %-24s %-12s %-20s %-20s\n",
					"ID", "USER", "NAME", "PREFIX", "LAST USED", "REVOKED")

				for _, k := range keys {
					_, _ = fmt.Fprintf(os.Stdout, "%-5d %-8d %-24s %-12s %-20s %-20s\n",
						k.ID, k.UserID, k.Name, k.Prefix, stamp(k.LastUsedAt), stamp(k.RevokedAt))
				}

				_ = sd.Shutdown()
			})
		},
	}
}

func revokeCommand() *cobra.Command {
	var id uint

	//nolint:exhaustruct_v5
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an API key; it stops working on the next request",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *apikeysvc.APIKeySvc, sd fx.Shutdowner, logger *zap.Logger) {
				if err := svc.Revoke(context.Background(), id); err != nil {
					logger.Error("failed to revoke api key", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				logger.Info("api key revoked", zap.Uint("id", id))

				_ = sd.Shutdown()
			})
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "api key id (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func stamp(t *time.Time) string {
	if t == nil {
		return "-"
	}

	return t.UTC().Format(time.DateTime)
}
