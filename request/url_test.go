package request_test

import (
	"testing"

	"github.com/1995parham/koochooloo/request"
)

func TestURLValidation(t *testing.T) {
	cases := []struct {
		url     string
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
	}

	for _, c := range cases {
		rq := request.URL{
			URL: c.url,
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
