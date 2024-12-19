package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/controllers"
	"github.com/mbvlabs/grafto/server/middleware"
)

func dashboardRoutes(router *echo.Echo, ctrl controllers.Dashboard, mw middleware.Middleware) {
	dashboardRouter := router.Group("/dashboard")

	dashboardRouter.GET("", func(c echo.Context) error {
		return ctrl.Index(c)
	}, mw.AuthOnly)
}
