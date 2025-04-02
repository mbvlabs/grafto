package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/routes/middleware"
	"github.com/mbvlabs/grafto/routes/paths"
)

func dashboardRoutes(
	router *echo.Echo,
	ctrl handlers.Dashboard,
) {
	router.GET(paths.DashboardHome.URL, func(c echo.Context) error {
		return ctrl.Index(c)
	}, middleware.AuthOnly).Name = paths.DashboardHome.Name
}
