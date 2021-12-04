package config

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/1995parham/koochooloo/internal/db"
	"github.com/1995parham/koochooloo/internal/logger"
	"github.com/1995parham/koochooloo/internal/metric"
	telemetry "github.com/1995parham/koochooloo/internal/telemetry/config"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

const (
	// Prefix indicates environment variables prefix.
	Prefix = "koochooloo_"
)

type (
	// Config holds all configurations.
	Config struct {
		Database   db.Config        `koanf:"database"`
		Monitoring metric.Config    `koanf:"monitoring"`
		Logger     logger.Config    `koanf:"logger"`
		Telemetry  telemetry.Config `koanf:"telemetry"`
	}
)

// New reads configuration with viper.
func New() Config {
	var instance Config

	k := koanf.New(".")

	// load default configuration from file
	if err := k.Load(structs.Provider(Default(), "koanf"), nil); err != nil {
		log.Fatalf("error loading default: %s", err)
	}

	// load configuration from file
	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		log.Printf("error loading config.yml: %s", err)
	}

	// load environment variables with given prefix.
	// logger__host_agent means host_agent in logger struct.
	if err := k.Load(env.Provider(Prefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, Prefix)), "__", ".")
	}), nil); err != nil {
		log.Printf("error loading environment variables: %s", err)
	}

	if err := k.Unmarshal("", &instance); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	indent, _ := json.MarshalIndent(instance, "", "\t")
	tmpl := `
	================ Loaded Configuration ================
	%s
	======================================================
	`
	log.Printf(tmpl, string(indent))

	return instance
}
