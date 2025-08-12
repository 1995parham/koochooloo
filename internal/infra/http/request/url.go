package request

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// URL represents short URL creation request.
type URL struct {
	URL    string     `json:"url"`
	Name   string     `json:"name"`
	Expire *time.Time `json:"expire"`
}

// Validate URL request.
func (r URL) Validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.URL, validation.Required, is.RequestURI),
		validation.Field(&r.Expire, validation.Min(time.Now())),
	); err != nil {
		return fmt.Errorf("url request validation failed: %w", err)
	}

	return nil
}
