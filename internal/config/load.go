package config

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/tidwall/pretty"
)

const (
	delimeter = "."
	seprator  = "__"

	// prefix indicates environment variables prefix.
	prefix = "koochooloo_"
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
		log.Printf("error loading config.toml: %s", err)
	}

	LoadEnv(k)

	var instance Config
	if err := k.Unmarshal("", &instance); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	indent, err := json.MarshalIndent(instance, "", "\t")
	if err != nil {
		panic(err)
	}

	indent = pretty.Color(indent, nil)
	tmpl := `
	================ Loaded Configuration ================
%s
	======================================================
	`
	log.Printf(tmpl, string(indent))

	return &instance
}

func LoadEnv(k *koanf.Koanf) {
	callback := func(source string) string {
		base := strings.ToLower(strings.TrimPrefix(source, prefix))

		return strings.ReplaceAll(base, seprator, delimeter)
	}

	// load environment variables
	if err := k.Load(env.Provider(prefix, delimeter, callback), nil); err != nil {
		log.Printf("error loading environment variables: %s", err)
	}
}
