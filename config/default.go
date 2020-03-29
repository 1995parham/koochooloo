package config

// Default return default configuration
// nolint: gomnd
func Default() Config {
	return Config{
		Debug: false,
		Database: Database{
			Name: "koochooloo",
			URL:  "mongodb://127.0.0.1:27017",
		},
		Monitoring: Monitoring{
			Address: ":8080",
			Enabled: true,
		},
	}
}
