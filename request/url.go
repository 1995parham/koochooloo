package request

import "time"

// URLReq represents short URL creation request
type URLReq struct {
	URL    string     `json:"url" validate:"required"`
	Name   string     `json:"name"`
	Expire *time.Time `json:"expire"`
}
