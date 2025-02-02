package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/views/paths"
)

func appRoutes(router *echo.Echo, handlers handlers.App) {
	router.GET("/", func(c echo.Context) error {
		return handlers.LandingPage(c)
	}).Name = paths.HomePage.Name()

	router.GET("/about", func(c echo.Context) error {
		return handlers.AboutPage(c)
	}).Name = paths.AboutPage.Name()
}
