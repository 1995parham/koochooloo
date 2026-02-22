package model

import "time"

// URL is a model for url with its attributes.
type URL struct {
	Key        string
	URL        string
	Count      int
	ExpireTime *time.Time
}

// IsExpired returns true if the URL has an expiration time that is in the past.
func (u URL) IsExpired() bool {
	return u.ExpireTime != nil && u.ExpireTime.Before(time.Now())
}
