package model

import "time"

// URL is a model for url with its attributes.
type URL struct {
	Key        string
	URL        string
	Count      int
	ExpireTime *time.Time
	// OwnerID is the id of the user who owns this short URL. It is nil for
	// anonymous shorts created through the public API.
	OwnerID *uint
}

// IsExpired returns true if the URL has an expiration time that is in the past.
func (u URL) IsExpired() bool {
	return u.ExpireTime != nil && u.ExpireTime.Before(time.Now())
}
