package web

import (
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/labstack/echo/v4"
)

func (w *Web) HomeRoutes() {
	w.router.GET("/", func(c echo.Context) error {
		return w.controllers.HomeIndex(c)
	})

	w.router.GET("/500", func(c echo.Context) error {
		telemetry.Logger.Info("context", "c", c.Request().Header)
		return w.controllers.InternalError(c)
	})
}
