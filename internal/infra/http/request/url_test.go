package request_test

import (
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/infra/http/request"
)

// nolint: funlen
func TestURLValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		url     string
		expire  time.Time
		isValid bool
	}{
		{
			url:     "",
			isValid: false,
		},
		{
			url:     "hello",
			isValid: false,
		},
		{
			url:     "hello.com",
			isValid: false,
		},
		{
			url:     "www.hello.com",
			isValid: false,
		},
		{
			url:     "http://www.hello.com",
			isValid: true,
		},
		{
			url:     "http://www.hello.com",
			expire:  time.Now().Add(time.Second),
			isValid: true,
		},
		{
			url:     "http://www.hello.com",
			expire:  time.Now().Add(-time.Second),
			isValid: false,
		},
	}

	for _, c := range cases {
		var expire *time.Time
		if !c.expire.IsZero() {
			expire = &c.expire
		}

		rq := request.URL{
			URL:    c.url,
			Expire: expire,
			Name:   "",
		}

		err := rq.Validate()

		if c.isValid && err != nil {
			t.Fatalf("valid request %+v has error %s", rq, err)
		}

		if !c.isValid && err == nil {
			t.Fatalf("invalid request %+v has no error", rq)
		}
	}
}
