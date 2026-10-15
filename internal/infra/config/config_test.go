package config_test

import (
	"testing"

	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/stretchr/testify/require"
)

const redacted = "<redacted>"

func TestRedactedHidesSecrets(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Database.Dialect = "postgres"
	cfg.Database.URL = "host=db user=u password=hunter2 dbname=k"
	cfg.Auth.JWTSecret = "jwt-secret"
	cfg.Auth.OIDC.ClientSecret = "client-secret"

	got := cfg.Redacted()

	require.Equal(t, redacted, got.Database.URL)
	require.Equal(t, redacted, got.Auth.JWTSecret)
	require.Equal(t, redacted, got.Auth.OIDC.ClientSecret)
	require.Equal(t, "jwt-secret", cfg.Auth.JWTSecret, "Redacted must not change the config it was called on")
}

func TestRedactedKeepsSQLitePathAndEmptySecrets(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Database.Dialect = "sqlite"
	cfg.Database.URL = "koochooloo.db"
	cfg.Auth.JWTSecret = ""
	cfg.Auth.OIDC.ClientSecret = ""

	got := cfg.Redacted()

	require.Equal(t, "koochooloo.db", got.Database.URL)
	require.Empty(t, got.Auth.JWTSecret, "an empty secret stays empty so a missing one is visible")
	require.Empty(t, got.Auth.OIDC.ClientSecret)
}
