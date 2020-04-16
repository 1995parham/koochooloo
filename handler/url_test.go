package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/handler"
	"github.com/1995parham/koochooloo/request"
	"github.com/1995parham/koochooloo/store"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type URLSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *URLSuite) SetupSuite() {
	suite.engine = echo.New()

	url := handler.URL{Store: store.NewMockURL()}
	url.Register(suite.engine.Group("/api"))
}

// nolint: funlen
func (suite *URLSuite) TestPostRetrieve() {
	cases := []struct {
		name     string
		code     int
		key      string
		url      string
		expire   time.Time
		retrieve int
	}{
		{
			name:     "Successful",
			code:     http.StatusOK,
			key:      "raha",
			url:      "https://elahe-dastan.github.io",
			expire:   time.Time{},
			retrieve: http.StatusFound,
		}, {
			name:   "Duplicate Key",
			code:   http.StatusBadRequest,
			key:    "raha",
			url:    "http://github.com",
			expire: time.Time{},
		}, {
			name:   "Invalid URL",
			code:   http.StatusBadRequest,
			key:    "parham",
			url:    "github.com",
			expire: time.Time{},
		}, {
			name:     "Automatic",
			code:     http.StatusOK,
			key:      "",
			url:      "https://google.com",
			expire:   time.Time{},
			retrieve: http.StatusFound,
		}, {
			name:     "Expire",
			code:     http.StatusOK,
			key:      "ex",
			url:      "https://instagram.com",
			expire:   time.Now().Add(-time.Minute),
			retrieve: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			var expire = &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			b, err := json.Marshal(request.URL{
				URL:    c.url,
				Name:   c.key,
				Expire: expire,
			})
			suite.NoError(err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/urls", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			suite.engine.ServeHTTP(w, req)
			suite.Equal(c.code, w.Code)

			if c.code == http.StatusOK {
				var resp string
				suite.NoError(json.NewDecoder(w.Body).Decode(&resp))

				if c.key != "" {
					suite.Equal("$"+c.key, resp)
				}

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/%s", resp), nil)

				suite.engine.ServeHTTP(w, req)
				suite.Equal(c.retrieve, w.Code)
			}
		})
	}
}

func TestURLSuite(t *testing.T) {
	suite.Run(t, new(URLSuite))
}
