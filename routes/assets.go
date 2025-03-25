package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func assetsRoutes(
	router *echo.Echo,
	handlers handlers.Assets,
) {
	router.GET(paths.Robots.URL, func(c echo.Context) error {
		return c.File("./resources/seo/robots.txt")
	}).Name = paths.Robots.Name

	router.GET(paths.Sitemap.URL, func(c echo.Context) error {
		return handlers.Sitemap(c)
	}).Name = paths.Sitemap.Name

	router.GET(paths.Favicon.URL, func(c echo.Context) error {
		return c.File("./static/images/favicon.ico")
	}).Name = paths.Favicon.Name
}
