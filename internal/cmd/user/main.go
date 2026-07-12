// Package user provides the `koochooloo user` command group for managing
// local accounts from the CLI (creating the first admin, listing, promoting
// and removing users).
package user

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/usersvc"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/userdb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/term"
)

// Register adds the user command group to the root command.
func Register(root *cobra.Command) {
	//nolint:exhaustruct
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage user accounts",
	}

	cmd.AddCommand(createCommand(), listCommand(), setRoleCommand(), deleteCommand())
	root.AddCommand(cmd)
}

// runFx builds the user-service fx graph and invokes fn against it.
func runFx(invoke any) {
	fx.New(
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(db.Provide),
		fx.Provide(fx.Annotate(userdb.ProvideDB, fx.As(new(userrepo.Repository)))),
		fx.Provide(usersvc.Provide),
		fx.NopLogger,
		fx.Invoke(invoke),
	).Run()
}

func createCommand() *cobra.Command {
	var (
		username   string
		admin      bool
		superadmin bool
	)

	//nolint:exhaustruct
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a local user account",
		Run: func(_ *cobra.Command, _ []string) {
			role := model.RoleUser

			switch {
			case superadmin:
				role = model.RoleSuperAdmin
			case admin:
				role = model.RoleAdmin
			}

			password, err := readPassword("Password: ")
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "reading password failed: %s\n", err)

				os.Exit(1)
			}

			runFx(func(svc *usersvc.UserSvc, sd fx.Shutdowner, logger *zap.Logger) {
				user, err := svc.Register(context.Background(), username, password, role)
				if err != nil {
					logger.Error("failed to create user", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				logger.Info("user created",
					zap.Uint("id", user.ID),
					zap.String("username", user.Username),
					zap.String("role", string(user.Role)),
				)

				_ = sd.Shutdown()
			})
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "username (required)")
	cmd.Flags().BoolVar(&admin, "admin", false, "grant the admin role")
	cmd.Flags().BoolVar(&superadmin, "superadmin", false, "grant the superadmin role")
	_ = cmd.MarkFlagRequired("username")

	return cmd
}

func listCommand() *cobra.Command {
	//nolint:exhaustruct
	return &cobra.Command{
		Use:   "list",
		Short: "List user accounts",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *usersvc.UserSvc, sd fx.Shutdowner, logger *zap.Logger) {
				users, err := svc.List(context.Background())
				if err != nil {
					logger.Error("failed to list users", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				_, _ = fmt.Fprintf(os.Stdout, "%-5s %-24s %-12s %-8s\n", "ID", "USERNAME", "ROLE", "PROVIDER")

				for _, u := range users {
					_, _ = fmt.Fprintf(os.Stdout, "%-5d %-24s %-12s %-8s\n", u.ID, u.Username, u.Role, u.Provider)
				}

				_ = sd.Shutdown()
			})
		},
	}
}

func setRoleCommand() *cobra.Command {
	var (
		id   uint
		role string
	)

	//nolint:exhaustruct
	cmd := &cobra.Command{
		Use:   "set-role",
		Short: "Change a user's role (user|admin|superadmin)",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *usersvc.UserSvc, sd fx.Shutdowner, logger *zap.Logger) {
				if err := svc.SetRole(context.Background(), id, model.Role(role)); err != nil {
					logger.Error("failed to set role", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				logger.Info("role updated", zap.Uint("id", id), zap.String("role", role))

				_ = sd.Shutdown()
			})
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "user id (required)")
	cmd.Flags().StringVar(&role, "role", "", "role: user|admin|superadmin (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func deleteCommand() *cobra.Command {
	var id uint

	//nolint:exhaustruct
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user account",
		Run: func(_ *cobra.Command, _ []string) {
			runFx(func(svc *usersvc.UserSvc, sd fx.Shutdowner, logger *zap.Logger) {
				if err := svc.Delete(context.Background(), id); err != nil {
					logger.Error("failed to delete user", zap.Error(err))

					_ = sd.Shutdown(fx.ExitCode(1))

					return
				}

				logger.Info("user deleted", zap.Uint("id", id))

				_ = sd.Shutdown()
			})
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "user id (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// readPassword reads a secret from the terminal without echoing. When stdin is
// not a terminal (piped input in scripts/CI) it reads a single line instead.
func readPassword(prompt string) (string, error) {
	_, _ = fmt.Fprint(os.Stdout, prompt)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		secret, err := term.ReadPassword(fd)

		_, _ = fmt.Fprintln(os.Stdout)

		if err != nil {
			return "", fmt.Errorf("reading terminal password: %w", err)
		}

		return string(secret), nil
	}

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading password line: %w", err)
	}

	return strings.TrimRight(line, "\r\n"), nil
}
