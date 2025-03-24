package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func dashboardRoutes(
	router *echo.Echo,
	ctrl handlers.Dashboard,
) {
	router.GET(paths.DashboardHome.URL, func(c echo.Context) error {
		return ctrl.Index(c)
	}, http.AuthOnly).Name = paths.DashboardHome.Name
}
