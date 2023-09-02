package telemetry

type Config struct {
	Trace       Trace  `json:"trace,omitempty"        koanf:"trace"`
	Meter       Meter  `json:"meter,omitempty"        koanf:"meter"`
	Namespace   string `json:"namespace,omitempty"    koanf:"namespace"`
	ServiceName string `json:"service_name,omitempty" koanf:"service_name"`
}

type Meter struct {
	Address string `json:"address,omitempty" koanf:"address"`
	Enabled bool   `json:"enabled,omitempty" koanf:"enabled"`
}

type Trace struct {
	Enabled  bool   `json:"enabled,omitempty"  koanf:"enabled"`
	Endpoint string `json:"endpoint,omitempty" koanf:"endpoint"`
}
