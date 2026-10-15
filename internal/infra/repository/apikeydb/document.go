package apikeydb

import (
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

// apiKeyRecord is the GORM persistence model for an API key.
type apiKeyRecord struct {
	ID         uint       `gorm:"column:id;primaryKey;autoIncrement"`
	UserID     uint       `gorm:"column:user_id;index;not null"`
	Name       string     `gorm:"column:name;size:255;not null"`
	Prefix     string     `gorm:"column:prefix;size:32;not null"`
	Hash       string     `gorm:"column:hash;uniqueIndex;size:64;not null"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	RevokedAt  *time.Time `gorm:"column:revoked_at"`
}

// TableName pins the table name so every dialect uses the same one.
func (apiKeyRecord) TableName() string {
	return "api_keys"
}

func toRecord(k model.APIKey) apiKeyRecord {
	return apiKeyRecord{
		ID:         k.ID,
		UserID:     k.UserID,
		Name:       k.Name,
		Prefix:     k.Prefix,
		Hash:       k.Hash,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
		RevokedAt:  k.RevokedAt,
	}
}

func toModel(r apiKeyRecord) model.APIKey {
	return model.APIKey{
		ID:         r.ID,
		UserID:     r.UserID,
		Name:       r.Name,
		Prefix:     r.Prefix,
		Hash:       r.Hash,
		CreatedAt:  r.CreatedAt,
		LastUsedAt: r.LastUsedAt,
		RevokedAt:  r.RevokedAt,
	}
}

func toModels(rs []apiKeyRecord) []model.APIKey {
	keys := make([]model.APIKey, len(rs))
	for i, r := range rs {
		keys[i] = toModel(r)
	}

	return keys
}
