package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/handler"
	"github.com/1995parham/koochooloo/internal/request"
	"github.com/1995parham/koochooloo/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type URLSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *URLSuite) SetupSuite() {
	suite.engine = echo.New()

	url := handler.URL{Store: store.NewMockURL(), Logger: zap.NewNop()}
	url.Register(suite.engine.Group("/api"))
}

func (suite *URLSuite) TestBadRequest() {
	require := suite.Require()

	// because there is no content-type header, request is categorized as a bad request.
	b, err := json.Marshal(request.URL{
		URL:    "https://elahe-dastan.github.io",
		Name:   "",
		Expire: nil,
	})
	require.NoError(err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/urls", bytes.NewReader(b))

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusBadRequest, w.Code)
}

func (suite *URLSuite) TestExpiration() {
	require := suite.Require()

	expire := time.Now().Add(time.Second)
	url := "https://instagram.com"
	key := "ex"

	b, err := json.Marshal(request.URL{
		URL:    url,
		Name:   key,
		Expire: &expire,
	})
	require.NoError(err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/urls", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var resp string

	require.NoError(json.NewDecoder(w.Body).Decode(&resp))
	require.Equal("$"+key, resp)

	time.Sleep(time.Second)

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/%s", resp), nil)

		suite.engine.ServeHTTP(w, req)
		require.Equal(http.StatusNotFound, w.Code)
	}
}

// nolint: funlen
func (suite *URLSuite) TestPostRetrieve() {
	require := suite.Require()

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
			name:     "Automatic Key Generation",
			code:     http.StatusOK,
			key:      "",
			url:      "https://google.com",
			expire:   time.Time{},
			retrieve: http.StatusFound,
		}, {
			name:   "Invalid Expiration",
			code:   http.StatusBadRequest,
			key:    "ex",
			url:    "https://instagram.com",
			expire: time.Now().Add(-time.Minute),
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			b, err := json.Marshal(request.URL{
				URL:    c.url,
				Name:   c.key,
				Expire: expire,
			})
			require.NoError(err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/urls", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			suite.engine.ServeHTTP(w, req)
			require.Equal(c.code, w.Code)

			if c.code == http.StatusOK {
				var resp string
				require.NoError(json.NewDecoder(w.Body).Decode(&resp))

				if c.key != "" {
					require.Equal("$"+c.key, resp)
				}

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/%s", resp), nil)

				suite.engine.ServeHTTP(w, req)
				require.Equal(c.retrieve, w.Code)
			}
		})
	}
}

func TestURLSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(URLSuite))
}
