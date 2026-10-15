package server

// Config holds HTTP server settings.
type Config struct {
	// AnonymousCreate lets anyone create short URLs through the public
	// `POST /api/urls`, without logging in. Turn it off on a host whose domain
	// should not hand out links for strangers (phishing); signed-in users and
	// API keys keep creating links through `POST /admin/api/urls`.
	AnonymousCreate bool `json:"anonymous_create" koanf:"anonymous_create"`
}
