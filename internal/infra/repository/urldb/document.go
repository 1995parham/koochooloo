package urldb

import (
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
)

// urlRecord is the GORM persistence model for a URL. The domain model stays
// free of ORM tags; mapping happens through toRecord/toModel.
type urlRecord struct {
	Key        string     `gorm:"column:key;primaryKey;size:255"`
	URL        string     `gorm:"column:url;not null"`
	Count      int        `gorm:"column:count;not null;default:0"`
	ExpireTime *time.Time `gorm:"column:expire_time;index"`
	OwnerID    *uint      `gorm:"column:owner_id;index"`
}

// TableName pins the table name so every dialect uses the same one.
func (urlRecord) TableName() string {
	return "urls"
}

func toRecord(u model.URL) urlRecord {
	return urlRecord{
		Key:        u.Key,
		URL:        u.URL,
		Count:      u.Count,
		ExpireTime: u.ExpireTime,
		OwnerID:    u.OwnerID,
	}
}

func toModel(r urlRecord) model.URL {
	return model.URL{
		Key:        r.Key,
		URL:        r.URL,
		Count:      r.Count,
		ExpireTime: r.ExpireTime,
		OwnerID:    r.OwnerID,
	}
}

func toModels(rs []urlRecord) []model.URL {
	urls := make([]model.URL, len(rs))
	for i, r := range rs {
		urls[i] = toModel(r)
	}

	return urls
}
