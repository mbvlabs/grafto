package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/controllers"
)

func appRoutes(router *echo.Echo, ctrl controllers.App) {
	router.GET("/", func(c echo.Context) error {
		return ctrl.LandingPage(c)
	})
}
