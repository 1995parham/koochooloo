package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/metric"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Debug: false,
		Database: db.Config{
			Name: "koochooloo",
			URL:  "mongodb://127.0.0.1:27017",
		},
		Monitoring: metric.Config{
			Address: ":8080",
			Enabled: true,
		},
	}
}
