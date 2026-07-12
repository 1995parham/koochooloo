// Package response defines the JSON payloads returned by the admin API. Using
// dedicated types keeps sensitive fields (e.g. password hashes) out of
// responses and decouples the wire format from the domain model.
package response

import (
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

// User is the public representation of a user account.
type User struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Role      model.Role `json:"role"`
	Provider  string     `json:"provider"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewUser maps a domain user onto its response form (dropping the password hash).
func NewUser(u model.User) User {
	return User{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		Provider:  string(u.Provider),
		CreatedAt: u.CreatedAt,
	}
}

// Users maps a slice of domain users.
func Users(us []model.User) []User {
	out := make([]User, len(us))
	for i, u := range us {
		out[i] = NewUser(u)
	}

	return out
}

// URL is the public representation of a short URL.
type URL struct {
	Key        string     `json:"key"`
	URL        string     `json:"url"`
	Count      int        `json:"count"`
	ExpireTime *time.Time `json:"expire_time"`
	OwnerID    *uint      `json:"owner_id"`
}

// NewURL maps a domain URL onto its response form.
func NewURL(u model.URL) URL {
	return URL{
		Key:        u.Key,
		URL:        u.URL,
		Count:      u.Count,
		ExpireTime: u.ExpireTime,
		OwnerID:    u.OwnerID,
	}
}

// URLs maps a slice of domain URLs.
func URLs(us []model.URL) []URL {
	out := make([]URL, len(us))
	for i, u := range us {
		out[i] = NewURL(u)
	}

	return out
}

// Token is the response to a successful login.
type Token struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
