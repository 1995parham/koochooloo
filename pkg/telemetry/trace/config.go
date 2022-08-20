package trace

type Config struct {
	Enabled bool    `koanf:"enabled"`
	Ratio   float64 `koanf:"ratio"`
	Host    string  `koanf:"host"`
	Port    int     `koanf:"port"`
}
