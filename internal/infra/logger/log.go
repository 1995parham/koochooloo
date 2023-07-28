package logger

import (
	"log"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level string `json:"level"  koanf:"level"`
}

// New creates a zap logger for console and also setup an output for syslog.
func Provide(lc fx.Lifecycle, cfg Config) *zap.Logger {
	var lvl zapcore.Level
	if err := lvl.Set(cfg.Level); err != nil {
		log.Printf("cannot parse log level %s: %s", cfg.Level, err)

		lvl = zapcore.WarnLevel
	}

	zapcfg := zap.NewDevelopmentConfig()
	zapcfg.Level.SetLevel(lvl)

	logger, err := zapcfg.Build()
	if err != nil {
		log.Fatalf("logger creation failed %s", err)
	}

	lc.Append(
		fx.StopHook(func() { _ = logger.Sync() }),
	)

	return logger
}
