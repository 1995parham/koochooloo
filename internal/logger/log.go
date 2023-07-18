package logger

import (
	"log"
	"log/syslog"
	"os"

	"github.com/tchap/zapext/v2/zapsyslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level  string `json:"level"  koanf:"level"`
	Syslog Syslog `json:"syslog" koanf:"syslog"`
}

type Syslog struct {
	Enabled bool   `json:"enabled" koanf:"enabled"`
	Network string `json:"network" koanf:"network"`
	Address string `json:"address" koanf:"address"`
	Tag     string `json:"tag"     koanf:"tag"`
}

// New creates a zap logger for console and also setup an output for syslog.
// nolint: nosnakecase
func New(cfg Config) *zap.Logger {
	var lvl zapcore.Level
	if err := lvl.Set(cfg.Level); err != nil {
		log.Printf("cannot parse log level %s: %s", cfg.Level, err)

		lvl = zapcore.WarnLevel
	}

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	defaultCore := zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(os.Stderr)), lvl)
	cores := []zapcore.Core{
		defaultCore,
	}

	if cfg.Syslog.Enabled {
		p := getPriorityFromLevel(lvl.String()) | syslog.LOG_LOCAL0
		encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

		writer, err := syslog.Dial(cfg.Syslog.Network, cfg.Syslog.Address, p, cfg.Syslog.Tag)
		if err == nil {
			cores = append(cores, zapsyslog.NewCore(lvl, encoder, writer))
		} else {
			log.Printf("cannot create syslog core, error: %s", err.Error())
			log.Println("warning, logger output is only stdout")
		}
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return logger
}

// nolint: nosnakecase
func getPriorityFromLevel(level string) syslog.Priority {
	switch level {
	case "debug":
		return syslog.LOG_DEBUG
	case "info":
		return syslog.LOG_INFO
	case "warn":
		return syslog.LOG_WARNING
	case "error":
		return syslog.LOG_ERR
	case "fatal":
		return syslog.LOG_CRIT
	case "panic":
		return syslog.LOG_ALERT
	default:
		return syslog.LOG_ERR
	}
}
