package auth

import (
	"fmt"
	"time"
)

// defaultTTL is used when the configured token_ttl is empty or unparseable.
const defaultTTL = 24 * time.Hour

// Provide builds a TokenService from configuration.
func Provide(cfg Config) (*TokenService, error) {
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("%w: jwt secret is empty", ErrInvalidToken)
	}

	ttl := defaultTTL

	if cfg.TokenTTL != "" {
		parsed, err := time.ParseDuration(cfg.TokenTTL)
		if err != nil {
			return nil, fmt.Errorf("parsing token_ttl %q: %w", cfg.TokenTTL, err)
		}

		ttl = parsed
	}

	return NewTokenService(cfg.JWTSecret, ttl), nil
}
