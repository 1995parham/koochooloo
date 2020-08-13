package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1995parham/koochooloo/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type HealthzSuite struct {
	suite.Suite

	engine *echo.Echo
}

func (suite *HealthzSuite) SetupSuite() {
	suite.engine = echo.New()

	handler.Healthz{}.Register(suite.engine.Group(""))
}

func (suite *HealthzSuite) TestHandler() {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	suite.engine.ServeHTTP(w, req)
	suite.Equal(http.StatusNoContent, w.Code)
}

func TestHealthzSuite(t *testing.T) {
	suite.Run(t, new(HealthzSuite))
}
