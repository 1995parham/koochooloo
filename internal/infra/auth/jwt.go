// Package auth issues and verifies the session JWTs that guard the admin API,
// and holds the authentication configuration (local + OIDC).
package auth

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken indicates a missing, malformed, expired or badly-signed token.
var ErrInvalidToken = errors.New("invalid token")

// alg is the only accepted JWT signing algorithm.
const alg = "HS256"

// Claims is the koochooloo session token payload.
type Claims struct {
	jwt.RegisteredClaims

	Username string     `json:"username"`
	Role     model.Role `json:"role"`
}

// TokenService issues and verifies session tokens.
type TokenService struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenService builds a token service from a secret and token lifetime.
func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

// Issue mints a signed token for the given user.
func (t *TokenService) Issue(user model.User, now time.Time) (string, error) {
	claims := ClaimsFor(user)
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(t.ttl))

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("signing token failed: %w", err)
	}

	return signed, nil
}

// ClaimsFor returns the identity claims of a user, without issue or expiry
// times. The API-key path uses it to give a key's request the same claims a
// session token for the key's owner would carry.
func ClaimsFor(user model.User) Claims {
	// modernize's embedlit wants the promoted fields hoisted into the outer
	// literal (a Go 1.27 feature), but exhaustruct v5.0.3 panics on any literal
	// that keys a promoted field, and a crashing analyzer cannot be nolint-ed.
	// Revisit once exhaustruct understands the Go 1.27 syntax.
	return Claims{
		RegisteredClaims: jwt.RegisteredClaims{ //nolint:exhaustruct_v5,modernize // only the subject identifies the user.
			Subject: strconv.FormatUint(uint64(user.ID), 10),
		},
		Username: user.Username,
		Role:     user.Role,
	}
}

// Parse verifies a token's signature and expiry and returns its claims.
func (t *TokenService) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{}, //nolint:exhaustruct_v5 // populated by the parser.
		func(_ *jwt.Token) (any, error) { return t.secret, nil },
		jwt.WithValidMethods([]string{alg}),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}

// UserID returns the numeric user id encoded in the token subject.
func (c Claims) UserID() (uint, error) {
	id, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil || id > math.MaxUint {
		return 0, fmt.Errorf("%w: bad subject", ErrInvalidToken)
	}

	return uint(id), nil
}
