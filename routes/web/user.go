package web

import "github.com/labstack/echo/v4"

func (w *Web) UserRoutes() {
	w.router.GET("/user/create", func(c echo.Context) error {
		return w.controllers.CreateUser(c)
	})
	w.router.POST("/user/store", func(c echo.Context) error {
		return w.controllers.StoreUser(c)
	})
	w.router.GET("/sign-in", func(c echo.Context) error {
		return w.controllers.AuthenticateUser(c)
	})
}
