package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
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
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
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
				Store:           store,
				Logger:          zap.NewNop(),
				Tracer:          tele.TraceProvider.Tracer(""),
				AnonymousCreate: true,
			}
			url.Register(suite.engine.Group("/api"))

			handler.Healthz{
				Logger: zap.NewNop(),
				Tracer: tele.TraceProvider.Tracer(""),
			}.Register(suite.engine.Group(""))

			url.RegisterRoot(suite.engine)
		}),
	).RequireStart().RequireStop()
}

func (suite *URLSuite) TestCountNotFound() {
	require := suite.Require()

	key := "notexists"

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/count/"+key, nil)
	require.NoError(err)

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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/urls", bytes.NewReader(b))
	require.NoError(err)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var resp string

	require.NoError(json.NewDecoder(w.Body).Decode(&resp))
	require.Equal("$"+key, resp)

	{
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/count/"+resp, nil)
		require.NoError(err)

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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/urls", bytes.NewReader(b))
	require.NoError(err)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusBadRequest, w.Code)
}

// TestExpiration runs inside a synctest bubble so the wait for the URL to
// expire costs no wall-clock time: time.Now and the sleep below both read the
// bubble's synthetic clock, which synctest.Sleep advances instantly.
func (suite *URLSuite) TestExpiration() {
	synctest.Test(suite.T(), func(t *testing.T) {
		require := require.New(t)

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
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/api/urls", bytes.NewReader(b))
		require.NoError(err)

		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		suite.engine.ServeHTTP(w, req)
		require.Equal(http.StatusOK, w.Code)

		var resp string

		require.NoError(json.NewDecoder(w.Body).Decode(&resp))
		require.Equal("$"+key, resp)

		// IsExpired compares with a strict Before, so step past the deadline
		// rather than exactly onto it.
		synctest.Sleep(2 * time.Second)

		{
			w := httptest.NewRecorder()
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/api/"+resp, nil)
			require.NoError(err)

			suite.engine.ServeHTTP(w, req)
			require.Equal(http.StatusNotFound, w.Code)
		}
	})
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
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/urls", bytes.NewReader(b))
			require.NoError(err)

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
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/"+resp, nil)
				require.NoError(err)

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

func (suite *URLSuite) TestRootRetrieveGenerated() {
	require := suite.Require()

	url := "https://hatchup.capital"
	key := suite.create(url, "")

	w := suite.get("/" + key)
	require.Equal(http.StatusFound, w.Code)
	require.Equal(url, w.Header().Get(echo.HeaderLocation))

	w = suite.get("/api/count/" + key)
	require.Equal(http.StatusOK, w.Code)
	require.Equal("1\n", w.Body.String())
}

func (suite *URLSuite) TestRootRetrieveAlias() {
	require := suite.Require()

	url := "https://team.hatchup.capital"
	key := suite.create(url, "team")
	require.Equal("$team", key)

	// The bare name, the form a short host hands out.
	w := suite.get("/team")
	require.Equal(http.StatusFound, w.Code)
	require.Equal(url, w.Header().Get(echo.HeaderLocation))

	// The stored "$" form keeps working.
	w = suite.get("/$team")
	require.Equal(http.StatusFound, w.Code)
	require.Equal(url, w.Header().Get(echo.HeaderLocation))
}

func (suite *URLSuite) TestRootRetrieveNotFound() {
	suite.Require().Equal(http.StatusNotFound, suite.get("/nothere").Code)
}

func (suite *URLSuite) TestRootDoesNotShadowStaticRoutes() {
	require := suite.Require()

	require.Equal(http.StatusNoContent, suite.get("/healthz").Code)

	// /api/<key> keeps resolving through its own route.
	key := suite.create("https://example.com", "")
	require.Equal(http.StatusFound, suite.get("/api/"+key).Code)
}

// create shortens url through the public API and returns the key it answers with.
func (suite *URLSuite) TestRetrieveEscapedAlias() {
	require := suite.Require()

	url := "https://hatchup.capital/escaped"
	key := suite.create(url, "escaped")
	require.Equal("$escaped", key)

	// A client that path-escapes the `$` of an alias reaches the same link.
	for _, path := range []string{"/api/%24escaped", "/%24escaped"} {
		w := suite.get(path)
		require.Equal(http.StatusFound, w.Code, path)
		require.Equal(url, w.Header().Get(echo.HeaderLocation), path)
	}
}

func (suite *URLSuite) create(url, name string) string {
	require := suite.Require()

	b, err := json.Marshal(request.URL{URL: url, Name: name, Expire: nil})
	require.NoError(err)

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/urls", bytes.NewReader(b))
	require.NoError(err)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var key string

	require.NoError(json.NewDecoder(w.Body).Decode(&key))

	return key
}

// get issues a GET for path on the suite's engine.
func (suite *URLSuite) get(path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	suite.Require().NoError(err)

	suite.engine.ServeHTTP(w, req)

	return w
}

func TestCreateAnonymousDisabled(t *testing.T) {
	t.Parallel()

	engine := echo.New()

	handler.URL{
		Store:           nil, // never reached: the flag is checked first.
		Logger:          zap.NewNop(),
		Tracer:          telemetry.ProvideNull(nil).TraceProvider.Tracer(""),
		AnonymousCreate: false,
	}.Register(engine.Group("/api"))

	body, err := json.Marshal(request.URL{URL: "https://example.com", Name: "", Expire: nil})
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/urls", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "anonymous link creation is disabled")
}
