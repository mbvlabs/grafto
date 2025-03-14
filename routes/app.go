package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func appRoutes(router *echo.Echo, handlers handlers.App) {
	router.GET("/", func(c echo.Context) error {
		return handlers.LandingPage(c)
	}).Name = paths.Home.String()

	router.GET("/about", func(c echo.Context) error {
		return handlers.AboutPage(c)
	}).Name = paths.About.String()

	router.GET("/redirect", func(c echo.Context) error {
		qp := c.QueryParam("to")
		return c.Redirect(http.StatusPermanentRedirect, qp)
	}).Name = paths.Redirect.String()
}
