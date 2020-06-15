package request

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// URL represents short URL creation request.
type URL struct {
	URL    string     `json:"url"`
	Name   string     `json:"name"`
	Expire *time.Time `json:"expire"`
}

// Validate URL request.
func (r URL) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.URL, validation.Required, is.RequestURI),
	)
}
