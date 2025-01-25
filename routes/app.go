package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
)

func appRoutes(router *echo.Echo, handlers handlers.App) {
	router.GET("/", func(c echo.Context) error {
		return handlers.LandingPage(c)
	})
}
