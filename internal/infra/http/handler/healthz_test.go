package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1995parham/koochooloo/internal/infra/http/handler"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type HealthzSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *HealthzSuite) SetupSuite() {
	suite.engine = echo.New()

	fxtest.New(suite.T(),
		fx.Provide(telemetry.ProvideNull),
		fx.Invoke(func(tele telemetry.Telemetery) {
			url := handler.Healthz{
				Logger: zap.NewNop(),
				Tracer: tele.TraceProvider.Tracer(""),
			}
			url.Register(suite.engine.Group(""))
		}),
	).RequireStart().RequireStop()
}

func (suite *HealthzSuite) TestHandler() {
	require := suite.Require()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	require.Equal(http.StatusNoContent, w.Code)
}

func TestHealthzSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HealthzSuite))
}
