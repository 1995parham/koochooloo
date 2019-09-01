package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/go-playground/validator.v9"
)

// App creates new instance of Echo and configures it
func App(debug bool) *echo.Echo {
	app := echo.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Pre(middleware.RemoveTrailingSlash())

	if debug {
		app.Logger.SetLevel(log.DEBUG)
	}

	// validator
	app.Validator = &DefaultValidator{validator.New()}

	return app
}
