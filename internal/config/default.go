package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/logger"
	telemetry "github.com/1995parham/koochooloo/internal/telemetry/config"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Logger: logger.Config{
			Level: "debug",
			Syslog: logger.Syslog{
				Enabled: false,
				Network: "",
				Address: "",
				Tag:     "",
			},
		},
		Database: db.Config{
			Name: "koochooloo",
			URL:  "mongodb://127.0.0.1:27017",
		},
		Telemetry: telemetry.Config{
			Namespace:   "1995parham",
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
