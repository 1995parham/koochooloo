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
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{ //nolint:exhaustruct // only the fields we set are relevant.
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
		Username: user.Username,
		Role:     user.Role,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("signing token failed: %w", err)
	}

	return signed, nil
}

// Parse verifies a token's signature and expiry and returns its claims.
func (t *TokenService) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{}, //nolint:exhaustruct // populated by the parser.
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
