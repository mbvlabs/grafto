package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/routes/paths"
)

func resourceRoutes(router *echo.Echo) {
	router.GET("/robots.txt", func(c echo.Context) error {
		return c.File("./resources/seo/robots.txt")
	}).Name = paths.Robots.String()

	router.GET("/sitemap.xml", func(c echo.Context) error {
		return c.File("./resources/seo/sitemap.xml")
	}).Name = paths.Sitemap.String()

	router.GET("/favicon.ico", func(c echo.Context) error {
		return c.File("./static/images/favicon.ico")
	}).Name = paths.Favicon.String()
}
