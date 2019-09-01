package models

import "time"

// URL is a model for url with its attributes
type URL struct {
	Key        string     `bson:"key"`
	URL        string     `bson:"url"`
	Count      int        `bson:"count"`
	ExpireTime *time.Time `bson:"expire_time"`
}
