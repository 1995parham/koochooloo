package metric

type Config struct {
	Enabled bool   `koanf:"enabled"`
	Host    string `koanf:"host"`
	Port    int    `koanf:"port"`
}
