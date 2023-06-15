package config

type Config struct {
	Trace       Trace  `koanf:"trace"`
	Meter       Meter  `koanf:"meter"`
	Namespace   string `koanf:"namespace"`
	ServiceName string `koanf:"service_name"`
}

type Meter struct {
	Address string `koanf:"address"`
	Enabled bool   `koanf:"enabled"`
}

type Trace struct {
	Enabled bool  `koanf:"enabled"`
	Agent   Agent `koanf:"agent"`
}

type Agent struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
}
