package urldb

import (
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

type urlDocument struct {
	Key        string     `bson:"key"`
	URL        string     `bson:"url"`
	Count      int        `bson:"count"`
	ExpireTime *time.Time `bson:"expire_time"`
}

func toDocument(u model.URL) urlDocument {
	return urlDocument{
		Key:        u.Key,
		URL:        u.URL,
		Count:      u.Count,
		ExpireTime: u.ExpireTime,
	}
}

func toModel(d urlDocument) model.URL {
	return model.URL{
		Key:        d.Key,
		URL:        d.URL,
		Count:      d.Count,
		ExpireTime: d.ExpireTime,
	}
}
