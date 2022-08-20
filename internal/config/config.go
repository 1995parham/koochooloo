package config

import (
	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/logger"
	"github.com/1995parham/koochooloo/internal/metric"
	telemetry "github.com/1995parham/koochooloo/internal/telemetry/config"
)

// Config holds all configurations.
type Config struct {
	Database   db.Config        `koanf:"database"`
	Monitoring metric.Config    `koanf:"monitoring"`
	Logger     logger.Config    `koanf:"logger"`
	Telemetry  telemetry.Config `koanf:"telemetry"`
}
