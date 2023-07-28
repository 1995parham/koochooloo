package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1995parham/koochooloo/internal/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type HealthzSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *HealthzSuite) SetupSuite() {
	suite.engine = echo.New()

	handler.Healthz{
		Logger: zap.NewNop(),
		Tracer: trace.NewNoopTracerProvider().Tracer(""),
	}.Register(suite.engine.Group(""))
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
