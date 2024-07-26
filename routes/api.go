package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/http/handlers"
)

func apiV1Routes(
	router *echo.Group,
	controllers handlers.Api,
) {
	router.GET("/health", func(c echo.Context) error {
		return controllers.AppHealth(c)
	})
}
