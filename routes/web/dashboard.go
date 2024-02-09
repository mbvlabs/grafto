package web

import (
	"github.com/labstack/echo/v4"
)

func (w *Web) DashboardRoutes() {
	w.router.GET("/dashboard", func(c echo.Context) error {
		return w.controllers.DashboardIndex(c)
	}, w.middleware.AuthOnly)
}
