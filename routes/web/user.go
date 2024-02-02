package web

import (
	"github.com/labstack/echo/v4"
)

func (w *Web) UserRoutes() {
	w.router.GET("/register", func(c echo.Context) error {
		return w.controllers.CreateUser(c)
	})
	w.router.POST("/register", func(c echo.Context) error {
		return w.controllers.StoreUser(c)
	})

	w.router.GET("/login", func(c echo.Context) error {
		return w.controllers.CreateAuthenticatedSession(c)
	})
	w.router.POST("/login", func(c echo.Context) error {
		return w.controllers.StoreAuthenticatedSession(c)
	})

	w.router.GET("/verify-email", func(c echo.Context) error {
		return w.controllers.VerifyEmail(c)
	})

	w.router.GET("/forgot-password", func(c echo.Context) error {
		return w.controllers.CreatePasswordReset(c)
	})
	w.router.POST("/forgot-password", func(c echo.Context) error {
		return w.controllers.StorePasswordReset(c)
	})
	w.router.GET("/reset-password", func(c echo.Context) error {
		return w.controllers.CreateResetPassword(c)
	})
	w.router.POST("/reset-password", func(c echo.Context) error {
		return w.controllers.StoreResetPassword(c)
	})
}
