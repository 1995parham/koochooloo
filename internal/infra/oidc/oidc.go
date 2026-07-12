// Package oidc integrates optional OpenID Connect login (e.g. Keycloak):
// building authorization URLs, verifying callback ID tokens and mapping
// provider claims onto koochooloo roles.
package oidc

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// discoveryTimeout bounds the provider discovery request at startup.
const discoveryTimeout = 10 * time.Second

var (
	// ErrDisabled is returned when an OIDC operation is attempted while OIDC is off.
	ErrDisabled = errors.New("oidc is not enabled")
	// errNoIDToken indicates the token response lacked an id_token.
	errNoIDToken = errors.New("id_token missing from token response")
	// errNonceMismatch indicates the id token nonce did not match.
	errNonceMismatch = errors.New("nonce mismatch")
)

// Identity is the federated identity resolved from a verified ID token.
type Identity struct {
	Subject  string
	Username string
	Role     model.Role
}

// Service performs the OIDC authorization-code flow and role mapping.
type Service struct {
	enabled  bool
	cfg      auth.OIDCConfig
	oauth2   oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// Provide builds the OIDC service. When OIDC is disabled it returns an inert
// service. A failed provider discovery is logged and leaves OIDC disabled
// rather than blocking server startup.
func Provide(cfg auth.Config, logger *zap.Logger) *Service {
	log := logger.Named("oidc")

	if !cfg.OIDC.Enabled {
		return &Service{enabled: false} //nolint:exhaustruct
	}

	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, cfg.OIDC.Issuer)
	if err != nil {
		log.Warn("oidc disabled: provider discovery failed",
			zap.String("issuer", cfg.OIDC.Issuer), zap.Error(err))

		return &Service{enabled: false} //nolint:exhaustruct
	}

	scopes := cfg.OIDC.Scopes
	if !slices.Contains(scopes, oidc.ScopeOpenID) {
		scopes = append([]string{oidc.ScopeOpenID}, scopes...)
	}

	log.Info("oidc enabled", zap.String("issuer", cfg.OIDC.Issuer))

	return &Service{
		enabled: true,
		cfg:     cfg.OIDC,
		oauth2: oauth2.Config{ //nolint:exhaustruct
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			RedirectURL:  cfg.OIDC.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       scopes,
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID}), //nolint:exhaustruct
	}
}

// Enabled reports whether OIDC login is available.
func (s *Service) Enabled() bool {
	return s.enabled
}

// AuthCodeURL returns the provider URL to redirect the user to, binding the
// login to the given CSRF state and replay-protection nonce.
func (s *Service) AuthCodeURL(state, nonce string) string {
	return s.oauth2.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Verify exchanges an authorization code, validates the returned ID token
// (including the nonce) and resolves the federated identity.
func (s *Service) Verify(ctx context.Context, code, expectedNonce string) (Identity, error) {
	if !s.enabled {
		return Identity{}, ErrDisabled
	}

	oauth2Token, err := s.oauth2.Exchange(ctx, code)
	if err != nil {
		return Identity{}, fmt.Errorf("code exchange failed: %w", err)
	}

	rawID, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return Identity{}, errNoIDToken
	}

	idToken, err := s.verifier.Verify(ctx, rawID)
	if err != nil {
		return Identity{}, fmt.Errorf("verifying id token failed: %w", err)
	}

	if idToken.Nonce != expectedNonce {
		return Identity{}, errNonceMismatch
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return Identity{}, fmt.Errorf("decoding claims failed: %w", err)
	}

	username := stringClaim(claims, "preferred_username", "email", "name")
	if username == "" {
		username = idToken.Subject
	}

	return Identity{
		Subject:  idToken.Subject,
		Username: username,
		Role:     s.mapRole(claims),
	}, nil
}

// mapRole maps the configured roles claim onto a koochooloo role, defaulting
// to the plain user role.
func (s *Service) mapRole(claims map[string]any) model.Role {
	if s.cfg.RolesClaim == "" {
		return model.RoleUser
	}

	values := claimValues(claims, s.cfg.RolesClaim)

	for _, v := range s.cfg.SuperAdminValues {
		if slices.Contains(values, v) {
			return model.RoleSuperAdmin
		}
	}

	for _, v := range s.cfg.AdminValues {
		if slices.Contains(values, v) {
			return model.RoleAdmin
		}
	}

	return model.RoleUser
}

// claimValues resolves a dotted claim path (e.g. "realm_access.roles") to the
// string values found there.
func claimValues(claims map[string]any, path string) []string {
	var current any = claims

	for part := range strings.SplitSeq(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current = object[part]
	}

	return toStringSlice(current)
}

// toStringSlice coerces a claim value (string or array) into a string slice.
func toStringSlice(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))

		for _, element := range typed {
			if s, ok := element.(string); ok {
				out = append(out, s)
			}
		}

		return out
	default:
		return nil
	}
}

// stringClaim returns the first non-empty string claim among keys.
func stringClaim(claims map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := claims[key].(string); ok && value != "" {
			return value
		}
	}

	return ""
}
