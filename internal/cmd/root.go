package cmd

import (
	"os"

	"github.com/1995parham/koochooloo/internal/cmd/migrate"
	"github.com/1995parham/koochooloo/internal/cmd/server"
	"github.com/1995parham/koochooloo/internal/config"
	"github.com/1995parham/koochooloo/internal/logger"
	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ExitFailure status code.
const ExitFailure = 1

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cfg := config.New()

	logger := logger.New(cfg.Logger)

	//nolint: exhaustruct
	root := &cobra.Command{
		Use:     "koochooloo",
		Short:   "Make your URLs shorter (smaller) and more memorable",
		Version: versioninfo.Short(),
	}

	server.Register(root, cfg, logger)
	migrate.Register(root, cfg, logger)

	if err := root.Execute(); err != nil {
		logger.Error("failed to execute root command", zap.Error(err))
		os.Exit(ExitFailure)
	}
}
