package userdb

import (
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

// userRecord is the GORM persistence model for a user account.
type userRecord struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement"`
	Username     string    `gorm:"column:username;uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"column:role;not null;default:user"`
	Provider     string    `gorm:"column:provider;not null;default:local"`
	Subject      string    `gorm:"column:subject;index"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName pins the table name so every dialect uses the same one.
func (userRecord) TableName() string {
	return "users"
}

func toRecord(u model.User) userRecord {
	return userRecord{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		Provider:     string(u.Provider),
		Subject:      u.Subject,
		CreatedAt:    u.CreatedAt,
	}
}

func toModel(r userRecord) model.User {
	return model.User{
		ID:           r.ID,
		Username:     r.Username,
		PasswordHash: r.PasswordHash,
		Role:         model.Role(r.Role),
		Provider:     model.Provider(r.Provider),
		Subject:      r.Subject,
		CreatedAt:    r.CreatedAt,
	}
}

func toModels(rs []userRecord) []model.User {
	users := make([]model.User, len(rs))
	for i, r := range rs {
		users[i] = toModel(r)
	}

	return users
}
