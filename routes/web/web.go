package web

import (
	"github.com/MBvisti/grafto/controllers"
	"github.com/labstack/echo/v4"
)

type Web struct {
	controllers controllers.Controller
	router      *echo.Echo
}

func NewWeb(router *echo.Echo, controllers controllers.Controller) Web {
	return Web{
		controllers,
		router,
	}
}

func (w *Web) SetupWebRoutes() {
	w.UtilityRoutes()

	w.HomeRoutes()
	w.UserRoutes()
	w.DashboardRoutes()
}
