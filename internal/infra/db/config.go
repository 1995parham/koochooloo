package db

// Config contains the database configuration.
type Config struct {
	Name string `json:"name" koanf:"name"`
	URL  string `json:"url"  koanf:"url"`
}
