package generator

type Generator interface {
	ShortURLKey() string
}

func Provide(cfg Config) Generator {
	// nolint: gocritic
	switch cfg.Type {
	case "simple":
		return new(Simple)
	}

	return new(Simple)
}
