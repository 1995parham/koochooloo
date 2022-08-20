package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/pkg/telemetry"
	"github.com/1995parham/koochooloo/pkg/telemetry/log"
	"github.com/1995parham/koochooloo/pkg/telemetry/metric"
	"github.com/1995parham/koochooloo/pkg/telemetry/trace"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Telemetry: &telemetry.Config{
			Log: &log.Config{
				Development: true,
				Encoding:    "console",
				Level:       "info",
			},
			Metric: &metric.Config{
				Enabled: false,
				Host:    "127.0.0.1",
				Port:    1234,
			},
			Trace: &trace.Config{
				Enabled: false,
				Ratio:   1.0,
				Host:    "127.0.0.1",
				Port:    1234,
			},
		},
		Database: &db.Config{
			Name: "koochooloo",
			URL:  "mongodb://127.0.0.1:27017",
		},
	}
}
