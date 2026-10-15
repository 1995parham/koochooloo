package db

// Config contains the database configuration.
type Config struct {
	// Dialect selects the database engine: "sqlite", "postgres", or "mysql".
	Dialect string `json:"dialect" koanf:"dialect"`
	// URL is the engine-specific data source name (DSN). For sqlite it is a
	// file path (or ":memory:"); for postgres/mysql it is a connection string.
	URL string `json:"url" koanf:"url"`
	// MigrateOnStart runs the schema migration (the `migrate` command) before
	// the server starts listening. Off by default; turn it on where nothing
	// else can run `migrate`, such as a distroless container.
	MigrateOnStart bool `json:"migrate_on_start" koanf:"migrate_on_start"`
}
