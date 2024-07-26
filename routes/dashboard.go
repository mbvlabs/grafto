package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/http/handlers"
	"github.com/mbv-labs/grafto/http/middleware"
)

func dashboardRoutes(router *echo.Echo, ctrl handlers.Dashboard, mw middleware.Middleware) {
	dashboardRouter := router.Group("/dashboard")

	dashboardRouter.GET("", func(c echo.Context) error {
		return ctrl.Index(c)
	}, mw.AuthOnly)
}
