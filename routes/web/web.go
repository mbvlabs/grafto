package web

import (
	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/labstack/echo/v4"
)

type Web struct {
	controllers controllers.Controller
	router      *echo.Echo
	middleware  middleware.Middleware
}

func NewWeb(router *echo.Echo, controllers controllers.Controller, middleware middleware.Middleware) Web {
	return Web{
		controllers,
		router,
		middleware,
	}
}

func (w *Web) SetupWebRoutes() {
	w.UtilityRoutes()
	w.LandingPageRoutes()
	w.AuthRoutes()
	w.DashboardRoutes()
}
