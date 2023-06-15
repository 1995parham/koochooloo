package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/logger"
	"github.com/1995parham/koochooloo/internal/telemetry"
)

// Config holds all configurations.
type Config struct {
	Database  db.Config        `koanf:"database"`
	Logger    logger.Config    `koanf:"logger"`
	Telemetry telemetry.Config `koanf:"telemetry"`
}
