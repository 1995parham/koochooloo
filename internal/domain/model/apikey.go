package model

import "time"

// APIKey lets a service call the admin API as the user that owns it, without
// an interactive login. Only a hash of the key is stored; the key itself is
// shown once, when it is created.
type APIKey struct {
	ID     uint
	UserID uint
	// Name says what the key is for (e.g. "hatchup-team").
	Name string
	// Prefix is the first characters of the key, kept to tell keys apart in
	// listings without storing the secret.
	Prefix string
	// Hash is the hex SHA-256 of the full key.
	Hash       string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}

// Revoked reports whether the key has been revoked.
func (k APIKey) Revoked() bool {
	return k.RevokedAt != nil
}
