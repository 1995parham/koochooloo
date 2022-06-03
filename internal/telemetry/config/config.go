package config

type Config struct {
	Trace Trace `koanf:"trace"`
}

type Trace struct {
	Enabled     bool   `koanf:"enabled"`
	Agent       Agent  `koanf:"agent"`
	Namespace   string `koanf:"namespace"`
	ServiceName string `koanf:"service_name"`
}

type Agent struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
}
