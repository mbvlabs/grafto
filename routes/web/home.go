package web

import (
	"github.com/labstack/echo/v4"
)

func (w *Web) HomeRoutes() {
	w.router.GET("/", func(c echo.Context) error {
		return w.controllers.HomeIndex(c)
	})
}
