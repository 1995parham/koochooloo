package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/pkg/telemetry"
)

// Config holds all configurations.
type Config struct {
	Database  *db.Config        `koanf:"database"`
	Telemetry *telemetry.Config `koanf:"telemetry"`
}
