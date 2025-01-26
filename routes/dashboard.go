package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/views/paths"
)

func dashboardRoutes(
	router *echo.Echo,
	ctrl handlers.Dashboard,
) {
	dashboardRouter := router.Group("/dashboard")

	dashboardRouter.GET("", func(c echo.Context) error {
		return ctrl.Index(c)
	}, http.AuthOnly).Name = paths.DashboardHomePage
}
