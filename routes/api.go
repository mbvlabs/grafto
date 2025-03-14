package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func apiV1Routes(
	router *echo.Group,
	handlers handlers.Api,
) {
	router.GET("/health", func(c echo.Context) error {
		return handlers.AppHealth(c)
	}).Name = paths.APIHealth.String()
}
