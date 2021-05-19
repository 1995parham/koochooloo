package db

// Database configuration.
type Config struct {
	Name string `koanf:"name"`
	URL  string `koanf:"url"`
}
