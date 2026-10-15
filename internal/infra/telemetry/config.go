package telemetry

type Config struct {
	Trace       Trace  `json:"trace"                  koanf:"trace"`
	Meter       Meter  `json:"meter"                  koanf:"meter"`
	Namespace   string `json:"namespace,omitempty"    koanf:"namespace"`
	ServiceName string `json:"service_name,omitempty" koanf:"service_name"`
}

// Meter exports metrics to Prometheus on Address when Enabled. When it is
// disabled, Stdout prints them instead (for local debugging); with both off
// metrics are dropped.
type Meter struct {
	Address string `json:"address,omitempty" koanf:"address"`
	Enabled bool   `json:"enabled,omitempty" koanf:"enabled"`
	Stdout  bool   `json:"stdout,omitempty"  koanf:"stdout"`
}

// Trace exports spans over OTLP/gRPC to Endpoint when Enabled. When it is
// disabled, Stdout pretty-prints every span instead (for local debugging);
// with both off spans are dropped.
type Trace struct {
	Enabled  bool   `json:"enabled,omitempty"  koanf:"enabled"`
	Endpoint string `json:"endpoint,omitempty" koanf:"endpoint"`
	Stdout   bool   `json:"stdout,omitempty"   koanf:"stdout"`
}
