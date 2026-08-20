package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/infra/auth"
	"github.com/1995parham/koochooloo/internal/infra/http/handler"
	"github.com/1995parham/koochooloo/internal/infra/http/middleware"
	"github.com/1995parham/koochooloo/internal/infra/http/response"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type VersionSuite struct {
	suite.Suite

	engine *echo.Echo
	tokens *auth.TokenService
}

func (suite *VersionSuite) SetupSuite() {
	suite.engine = echo.New()
	suite.tokens = auth.NewTokenService("secret", time.Hour)

	fxtest.New(suite.T(),
		fx.Provide(telemetry.ProvideNull),
		fx.Invoke(func(tele telemetry.Telemetery) {
			version := handler.Version{
				Logger: zap.NewNop(),
				Tracer: tele.TraceProvider.Tracer(""),
			}

			authMw := middleware.Auth{Tokens: suite.tokens}
			version.Register(suite.engine.Group("", authMw.Authenticate))
		}),
	).RequireStart().RequireStop()
}

// The version is always reported; tests run from a working tree so it is the
// "devel" placeholder rather than a tag.
func (suite *VersionSuite) TestVersionAlwaysPresent() {
	suite.Require().NotEmpty(suite.get(model.RoleUser).Version)
	suite.Require().NotEmpty(suite.get(model.RoleAdmin).Version)
}

// Build provenance is admin-only.
func (suite *VersionSuite) TestRevisionHiddenFromUsers() {
	rs := suite.get(model.RoleUser)

	suite.Require().Empty(rs.Revision)
	suite.Require().Nil(rs.LastCommit)
	suite.Require().False(rs.Dirty)
}

func (suite *VersionSuite) TestUnauthenticated() {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/version", nil)

	suite.engine.ServeHTTP(w, req)
	suite.Require().Equal(http.StatusUnauthorized, w.Code)
}

// get calls the version endpoint as a user with the given role.
func (suite *VersionSuite) get(role model.Role) response.Version {
	require := suite.Require()

	//nolint: exhaustruct_v5
	token, err := suite.tokens.Issue(model.User{ID: 1, Username: "u", Role: role}, time.Now())
	require.NoError(err)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/version", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusOK, w.Code)

	var rs response.Version
	require.NoError(json.Unmarshal(w.Body.Bytes(), &rs))

	return rs
}

func TestVersionSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(VersionSuite))
}
