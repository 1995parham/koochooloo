package config

import (
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"go.uber.org/fx"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Out: fx.Out{},
		Logger: logger.Config{
			Level: "debug",
		},
		Generator: generator.Config{
			Type: "simple",
		},
		Database: db.Config{
			Name: "koochooloo",
			URL:  "mongodb://127.0.0.1:27017",
		},
		Telemetry: telemetry.Config{
			Namespace:   "1995parham.me",
			ServiceName: "koochooloo",
			Meter: telemetry.Meter{
				Address: ":8080",
				Enabled: true,
			},
			Trace: telemetry.Trace{
				Enabled:  true,
				Endpoint: "127.0.0.1:4317",
			},
		},
	}
}
