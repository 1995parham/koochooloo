package config

import (
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Logger: logger.Config{
			Level: "debug",
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
				Enabled: false,
				Agent: telemetry.Agent{
					Port: "6831",
					Host: "127.0.0.1",
				},
			},
		},
	}
}
