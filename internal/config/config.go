package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/logger"
	"github.com/1995parham/koochooloo/internal/telemetry"
)

// Config holds all configurations.
type Config struct {
	Database  db.Config        `json:"database"  koanf:"database"`
	Logger    logger.Config    `json:"logger"    koanf:"logger"`
	Telemetry telemetry.Config `json:"telemetry" koanf:"telemetry"`
}
