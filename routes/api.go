package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
)

func apiV1Routes(
	router *echo.Group,
	handlers handlers.Api,
) {
	router.GET("/health", func(c echo.Context) error {
		return handlers.AppHealth(c)
	})
}
