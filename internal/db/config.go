package db

// Config contains the database configuration.
type Config struct {
	Name string `koanf:"name"`
	URL  string `koanf:"url"`
}
