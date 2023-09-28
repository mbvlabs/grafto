package web

import (
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/labstack/echo/v4"
)

func (w *Web) DashboardRoutes() {
	w.router.GET("/dashboard", func(c echo.Context) error {
		return w.controllers.DashboardIndex(c)
	}, middleware.AuthOnly)
}
