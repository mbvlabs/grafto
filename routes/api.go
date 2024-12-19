package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/controllers"
)

func apiV1Routes(
	router *echo.Group,
	controllers controllers.Api,
) {
	router.GET("/health", func(c echo.Context) error {
		return controllers.AppHealth(c)
	})
}
