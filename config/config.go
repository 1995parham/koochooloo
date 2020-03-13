package config

import (
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds all configurations
type Config struct {
	Debug    bool
	Database struct {
		Name string
		URL  string
	}
}

// New reads configuration with viper
func New() Config {
	var instance Config

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigName("config.default")

	if err := v.ReadConfig(bytes.NewBufferString(Default)); err != nil {
		logrus.Fatalf("fatal error loading **default** config file: %s \n", err)
	}

	v.SetConfigName("config")

	if err := v.MergeInConfig(); err != nil {
		logrus.Warnf("no config file found, using defaults and environment variables")
	}

	v.SetEnvPrefix("urlshortener")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.UnmarshalExact(&instance); err != nil {
		logrus.Fatalf("unmarshaling error: %s", err)
	}

	logrus.Infof("following configuration is loaded:\n%+v", instance)

	return instance
}
