package config

type Config struct {
	Trace `koanf:"trace"`
}

type Trace struct {
	Enabled bool   `koanf:"enabled"`
	URL     string `koanf:"url"`
}
