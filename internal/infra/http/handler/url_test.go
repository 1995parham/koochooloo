package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/db"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/1995parham/koochooloo/internal/infra/http/handler"
	"github.com/1995parham/koochooloo/internal/infra/http/request"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type URLSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *URLSuite) SetupSuite() {
	suite.engine = echo.New()

	fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(db.Provide),
		fx.Provide(generator.Provide),
		fx.Provide(urlsvc.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(
			fx.Annotate(urldb.ProvideMemory, fx.As(new(urlrepo.Repository))),
		),
		fx.Invoke(func(store *urlsvc.URLSvc, tele telemetry.Telemetery) {
			url := handler.URL{
				Store:  store,
				Logger: zap.NewNop(),
				Tracer: tele.TraceProvider.Tracer(""),
			}
			url.Register(suite.engine.Group("/api"))
		}),
	).RequireStart().RequireStop()
}

func (suite *URLSuite) TestCountNotFound() {
	require := suite.Require()

	key := "notexists"

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/count/"+key, nil)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusNotFound, w.Code)
}

func (suite *URLSuite) TestCount() {
	require := suite.Require()

	url := "https://irandoc.ir"
	key := "doc"

	b, err := json.Marshal(request.URL{
		URL:    url,
		Name:   key,
		Expire: nil,
	})
	require.NoError(err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var resp string

	require.NoError(json.NewDecoder(w.Body).Decode(&resp))
	require.Equal("$"+key, resp)

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/count/"+resp, nil)

		suite.engine.ServeHTTP(w, req)
		require.Equal(http.StatusOK, w.Code)

		var count int

		require.NoError(json.NewDecoder(w.Body).Decode(&count))
		require.Equal(0, count)
	}
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
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(b))

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
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var resp string

	require.NoError(json.NewDecoder(w.Body).Decode(&resp))
	require.Equal("$"+key, resp)

	time.Sleep(time.Second)

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/"+resp, nil)

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
			name:     "Duplicate Key",
			code:     http.StatusBadRequest,
			key:      "raha",
			url:      "http://github.com",
			expire:   time.Time{},
			retrieve: 0,
		}, {
			name:     "Invalid URL",
			code:     http.StatusBadRequest,
			key:      "parham",
			url:      "github.com",
			expire:   time.Time{},
			retrieve: 0,
		}, {
			name:     "Automatic Key Generation",
			code:     http.StatusOK,
			key:      "",
			url:      "https://google.com",
			expire:   time.Time{},
			retrieve: http.StatusFound,
		}, {
			name:     "Invalid Expiration",
			code:     http.StatusBadRequest,
			key:      "ex",
			url:      "https://instagram.com",
			expire:   time.Now().Add(-time.Minute),
			retrieve: 0,
		},
	}

	for _, c := range cases {
		suite.Run(c.name, func() {
			expire := new(time.Time)

			*expire = c.expire
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
			req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(b))
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
				req := httptest.NewRequest(http.MethodGet, "/api/"+resp, nil)

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
