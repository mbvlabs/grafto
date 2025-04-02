package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func appRoutes(router *echo.Echo, handlers handlers.App) {
	router.GET(paths.LandingPage.URL, func(c echo.Context) error {
		return handlers.LandingPage(c)
	}).Name = paths.LandingPage.Name

	router.GET(paths.About.URL, func(c echo.Context) error {
		return handlers.AboutPage(c)
	}).Name = paths.About.Name

	router.GET(paths.Redirect.URL, func(c echo.Context) error {
		return handlers.Redirect(c)
	}).Name = paths.Redirect.Name
}
