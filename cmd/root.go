package cmd

import (
	"os"

	"github.com/1995parham/koochooloo/cmd/migrate"
	"github.com/1995parham/koochooloo/cmd/server"
	"github.com/1995parham/koochooloo/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ExitFailure status code
const ExitFailure = 1

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cfg := config.New()

	var root = &cobra.Command{
		Use:   "koochooloo",
		Short: "Make your URLs shorter (smaller) and more memorable",
	}

	server.Register(root, cfg)
	migrate.Register(root, cfg)

	if err := root.Execute(); err != nil {
		logrus.Errorf("failed to execute root command: %s", err.Error())
		os.Exit(ExitFailure)
	}
}
