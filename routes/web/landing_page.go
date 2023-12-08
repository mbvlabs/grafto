package web

import (
	"github.com/labstack/echo/v4"
)

func (w *Web) LandingPageRoutes() {
	w.router.GET("/", func(c echo.Context) error {
		return w.controllers.LandingPage(c)
	})
}
