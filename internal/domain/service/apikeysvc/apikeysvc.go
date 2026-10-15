// Package apikeysvc contains API-key business logic: minting a key for a
// user, authenticating a presented key, listing and revoking keys.
package apikeysvc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/apikeyrepo"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
)

const (
	// Prefix starts every key, so a bearer value is told apart from a session
	// JWT without parsing it, and a leaked key is easy to recognise.
	Prefix = "kch_"
	// secretLength is the number of random base62 characters after Prefix.
	secretLength = 40
	// displayLength is how much of the key is kept in clear for listings.
	displayLength = len(Prefix) + 6
	// touchEvery bounds how often last_used_at is written for a busy key.
	touchEvery = time.Minute

	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	// ErrInvalidKey indicates a missing, malformed, unknown or revoked key.
	ErrInvalidKey = errors.New("invalid api key")
	// ErrEmptyName indicates a key was requested without a name.
	ErrEmptyName = errors.New("api key name is required")
)

// APIKeySvc is the API-key service.
type APIKeySvc struct {
	keys  apikeyrepo.Repository
	users userrepo.Repository
	now   func() time.Time
}

// Provide creates a new API-key service.
func Provide(keys apikeyrepo.Repository, users userrepo.Repository) *APIKeySvc {
	return &APIKeySvc{keys: keys, users: users, now: time.Now}
}

// Create mints a key for the user with the given username. It returns the key
// in clear exactly once; only its hash is stored.
func (s *APIKeySvc) Create(ctx context.Context, username, name string) (string, model.APIKey, error) {
	if strings.TrimSpace(name) == "" {
		return "", model.APIKey{}, ErrEmptyName
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return "", model.APIKey{}, fmt.Errorf("finding user failed: %w", err)
	}

	secret, err := randomSecret()
	if err != nil {
		return "", model.APIKey{}, err
	}

	plain := Prefix + secret

	key, err := s.keys.Create(ctx, model.APIKey{
		ID:         0,
		UserID:     user.ID,
		Name:       strings.TrimSpace(name),
		Prefix:     plain[:displayLength],
		Hash:       hash(plain),
		CreatedAt:  s.now(),
		LastUsedAt: nil,
		RevokedAt:  nil,
	})
	if err != nil {
		return "", model.APIKey{}, fmt.Errorf("creating api key failed: %w", err)
	}

	return plain, key, nil
}

// Authenticate resolves a presented key to the user that owns it. Any failure
// - wrong shape, unknown, revoked, or an owner that no longer exists - is
// reported as ErrInvalidKey so a caller cannot tell them apart.
func (s *APIKeySvc) Authenticate(ctx context.Context, plain string) (model.User, error) {
	key, err := s.lookup(ctx, plain)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.users.FindByID(ctx, key.UserID)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return model.User{}, ErrInvalidKey
		}

		return model.User{}, fmt.Errorf("looking up api key owner failed: %w", err)
	}

	s.touch(ctx, key)

	return user, nil
}

// List returns every key, revoked ones included.
func (s *APIKeySvc) List(ctx context.Context) ([]model.APIKey, error) {
	keys, err := s.keys.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing api keys failed: %w", err)
	}

	return keys, nil
}

// Revoke revokes the key with the given id. It takes effect on the next request.
func (s *APIKeySvc) Revoke(ctx context.Context, id uint) error {
	if err := s.keys.Revoke(ctx, id, s.now()); err != nil {
		return fmt.Errorf("revoking api key failed: %w", err)
	}

	return nil
}

// lookup finds the live key matching plain.
func (s *APIKeySvc) lookup(ctx context.Context, plain string) (model.APIKey, error) {
	if !strings.HasPrefix(plain, Prefix) || len(plain) != len(Prefix)+secretLength {
		return model.APIKey{}, ErrInvalidKey
	}

	digest := hash(plain)

	key, err := s.keys.FindByHash(ctx, digest)
	if err != nil {
		if errors.Is(err, apikeyrepo.ErrAPIKeyNotFound) {
			return model.APIKey{}, ErrInvalidKey
		}

		return model.APIKey{}, fmt.Errorf("looking up api key failed: %w", err)
	}

	// The lookup is by hash, so the secret is never compared directly; the
	// constant-time check guards against a store that matches loosely.
	if subtle.ConstantTimeCompare([]byte(key.Hash), []byte(digest)) != 1 || key.Revoked() {
		return model.APIKey{}, ErrInvalidKey
	}

	return key, nil
}

// touch records the use, at most once per touchEvery so a busy key does not
// write on every request. It is best effort: bookkeeping must not fail a request.
func (s *APIKeySvc) touch(ctx context.Context, key model.APIKey) {
	now := s.now()
	if key.LastUsedAt != nil && now.Sub(*key.LastUsedAt) < touchEvery {
		return
	}

	_ = s.keys.Touch(ctx, key.ID, now)
}

// hash returns the hex SHA-256 of a key. A fast hash is enough here: the key
// carries 238 bits of randomness, so there is nothing to brute-force.
func hash(plain string) string {
	sum := sha256.Sum256([]byte(plain))

	return hex.EncodeToString(sum[:])
}

func randomSecret() (string, error) {
	limit := big.NewInt(int64(len(alphabet)))

	var b strings.Builder

	b.Grow(secretLength)

	for range secretLength {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generating api key failed: %w", err)
		}

		b.WriteByte(alphabet[n.Int64()])
	}

	return b.String(), nil
}
