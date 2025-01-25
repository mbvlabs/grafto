package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http"
	"github.com/mbvlabs/grafto/http/handlers"
)

func dashboardRoutes(
	router *echo.Echo,
	ctrl handlers.Dashboard,
) {
	dashboardRouter := router.Group("/dashboard")

	dashboardRouter.GET("", func(c echo.Context) error {
		return ctrl.Index(c)
	}, http.AuthOnly)
}
