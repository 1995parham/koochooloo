package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

const (
	delimeter = "."
	seperator = "__"

	// prefix indicates environment variables prefix.
	prefix = "koochooloo_"

	upTemplate     = "================ Loaded Configuration ================"
	bottomTemplate = "======================================================"
)

// New reads configuration with koanf.
func New() *Config {

	k := koanf.New(".")

	// load default configuration from default function
	if err := k.Load(structs.Provider(Default(), "koanf"), nil); err != nil {
		log.Fatalf("error loading default: %s", err)
	}

	// load configuration from file
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Printf("error loading config.yml: %s", err)
	}

	if err := LoadEnv(k); err != nil {
		log.Fatalf("error loading default values: %v", err)
	}

	var instance Config
	if err := k.Unmarshal("", &instance); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	log.Printf("%s\n%v\n%s\n", upTemplate, spew.Sdump(instance), bottomTemplate)

	return &instance
}

func LoadEnv(k *koanf.Koanf) error {
	callback := func(source string) string {
		base := strings.ToLower(strings.TrimPrefix(source, prefix))
		return strings.ReplaceAll(base, seperator, delimeter)
	}

	// load environment variables
	if err := k.Load(env.Provider(prefix, delimeter, callback), nil); err != nil {
		return fmt.Errorf("error loading environment variables: %s", err)
	}

	return nil
}
