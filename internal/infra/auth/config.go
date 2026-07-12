package auth

// Config holds authentication settings.
type Config struct {
	// JWTSecret signs and verifies session tokens (HMAC-SHA256). Override it
	// in production; the default is intended only for local development.
	JWTSecret string `json:"jwt_secret" koanf:"jwt_secret"`
	// TokenTTL is how long an issued token stays valid, as a Go duration
	// string (e.g. "24h", "30m").
	TokenTTL string `json:"token_ttl" koanf:"token_ttl"`
	// OIDC configures optional federated login (e.g. Keycloak).
	OIDC OIDCConfig `json:"oidc" koanf:"oidc"`
}

// OIDCConfig configures optional OpenID Connect login.
type OIDCConfig struct {
	// Enabled turns OIDC login on. When false the provider is not initialised.
	Enabled bool `json:"enabled" koanf:"enabled"`
	// Issuer is the provider's issuer URL (e.g. https://kc/realms/koochooloo).
	Issuer string `json:"issuer" koanf:"issuer"`
	// ClientID and ClientSecret are the registered OIDC client credentials.
	ClientID     string `json:"client_id"     koanf:"client_id"`
	ClientSecret string `json:"client_secret" koanf:"client_secret"`
	// RedirectURL is this app's callback URL registered with the provider.
	RedirectURL string `json:"redirect_url" koanf:"redirect_url"`
	// Scopes requested from the provider ("openid" is always included).
	Scopes []string `json:"scopes" koanf:"scopes"`
	// RolesClaim is the ID-token claim inspected for role mapping. Dotted
	// paths are supported (e.g. "realm_access.roles" for Keycloak).
	RolesClaim string `json:"roles_claim" koanf:"roles_claim"`
	// AdminValues / SuperAdminValues are claim values that grant the admin or
	// superadmin role; anything else maps to the plain user role.
	AdminValues      []string `json:"admin_values"      koanf:"admin_values"`
	SuperAdminValues []string `json:"superadmin_values" koanf:"superadmin_values"`
}
