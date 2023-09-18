package web

import (
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/labstack/echo/v4"
)

func (w *Web) UserRoutes() {
	w.router.GET("/user/create", func(c echo.Context) error {
		return w.controllers.CreateUser(c)
	}, middleware.CsrfMiddlewareFunc)
	w.router.POST("/user/store", func(c echo.Context) error {
		return w.controllers.StoreUser(c)
	}, middleware.CsrfMiddlewareFunc)
	w.router.GET("/sign-in", func(c echo.Context) error {
		return w.controllers.AuthenticateUser(c)
	})
}
