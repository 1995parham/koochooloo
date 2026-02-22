package generator

import (
	domgen "github.com/1995parham/koochooloo/internal/domain/generator"
)

func Provide(cfg Config) domgen.Generator {
	// nolint: gocritic
	switch cfg.Type {
	case "simple":
		return new(Simple)
	}

	return new(Simple)
}
