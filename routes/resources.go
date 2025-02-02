package routes

import (
	"github.com/labstack/echo/v4"
)

func resourceRoutes(router *echo.Echo) {
	router.GET("/robots.txt", func(c echo.Context) error {
		return c.File("./resources/seo/robots.txt")
	})
	router.GET("/sitemap.xml", func(c echo.Context) error {
		return c.File("./resources/seo/sitemap.xml")
	})
	router.GET("/favicon.ico", func(c echo.Context) error {
		return c.File("./static/images/favicon.ico")
	})
}
