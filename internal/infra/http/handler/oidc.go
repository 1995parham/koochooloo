package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

const (
	oidcStateCookie = "koochooloo_oidc_state"
	oidcNonceCookie = "koochooloo_oidc_nonce"
	oidcCookieTTL   = 5 * time.Minute
	// frontendLogin is where the browser lands after the flow; the session
	// token (or an error) is passed in the URL fragment so it never reaches a
	// server log.
	frontendLogin = "/admin/"
)

// AuthInfo advertises which login methods are available to the SPA.
func (h Auth) AuthInfo(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"oidc_enabled":   h.Provider.Enabled(),
		"oidc_login_url": "/admin/api/auth/oidc/login",
	})
}

// OIDCLogin starts the authorization-code flow: it stores a CSRF state and a
// replay nonce in short-lived cookies and redirects to the provider.
func (h Auth) OIDCLogin(c *echo.Context) error {
	if !h.Provider.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "oidc is not enabled")
	}

	state, err := randomToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "state generation failed")
	}

	nonce, err := randomToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "nonce generation failed")
	}

	h.setTempCookie(c, oidcStateCookie, state)
	h.setTempCookie(c, oidcNonceCookie, nonce)

	return c.Redirect(http.StatusFound, h.Provider.AuthCodeURL(state, nonce))
}

// OIDCCallback completes the flow: it validates the state and nonce, verifies
// the ID token, provisions/updates the user and redirects to the SPA with a
// freshly issued session token in the URL fragment.
func (h Auth) OIDCCallback(c *echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.auth.oidc_callback")
	defer span.End()

	if !h.Provider.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "oidc is not enabled")
	}

	state, err := c.Cookie(oidcStateCookie)
	if err != nil || state.Value == "" || state.Value != c.QueryParam("state") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid oidc state")
	}

	nonce, err := c.Cookie(oidcNonceCookie)
	if err != nil || nonce.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing oidc nonce")
	}

	h.clearTempCookie(c, oidcStateCookie)
	h.clearTempCookie(c, oidcNonceCookie)

	identity, err := h.Provider.Verify(ctx, c.QueryParam("code"), nonce.Value)
	if err != nil {
		span.RecordError(err)
		h.Logger.Warn("oidc verification failed", zap.Error(err))

		return c.Redirect(http.StatusFound, frontendLogin+"#error=oidc_failed")
	}

	user, err := h.Users.EnsureOIDC(ctx, identity.Subject, identity.Username, identity.Role)
	if err != nil {
		span.RecordError(err)
		h.Logger.Error("oidc user provisioning failed", zap.Error(err))

		return c.Redirect(http.StatusFound, frontendLogin+"#error=provisioning_failed")
	}

	token, err := h.Tokens.Issue(user, time.Now())
	if err != nil {
		span.RecordError(err)

		return c.Redirect(http.StatusFound, frontendLogin+"#error=token_failed")
	}

	return c.Redirect(http.StatusFound, frontendLogin+"#token="+url.QueryEscape(token))
}

// setTempCookie stores a short-lived, http-only cookie scoped to the auth path.
func (h Auth) setTempCookie(c *echo.Context, name, value string) {
	c.SetCookie(&http.Cookie{ //nolint:exhaustruct,gosec // Secure is set from TLS; local dev serves over http.
		Name:     name,
		Value:    value,
		Path:     "/admin/api/auth",
		MaxAge:   int(oidcCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearTempCookie expires a cookie set by setTempCookie.
func (h Auth) clearTempCookie(c *echo.Context, name string) {
	c.SetCookie(&http.Cookie{ //nolint:exhaustruct,gosec // Secure is set from TLS; local dev serves over http.
		Name:     name,
		Value:    "",
		Path:     "/admin/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// randomToken returns a URL-safe 256-bit random string.
func randomToken() (string, error) {
	buf := make([]byte, 32) //nolint:mnd
	if _, err := rand.Read(buf); err != nil {
		return "", err //nolint:wrapcheck // caller wraps into an HTTP error.
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
