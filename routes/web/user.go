package web

import (
	"github.com/labstack/echo/v4"
)

func (w *Web) UserRoutes() {
	w.router.GET("/user/create", func(c echo.Context) error {
		return w.controllers.CreateUser(c)
	})
	w.router.POST("/user/store", func(c echo.Context) error {
		return w.controllers.StoreUser(c)
	})
	w.router.GET("/login", func(c echo.Context) error {
		return w.controllers.Login(c)
	})
	w.router.POST("/authenticate", func(c echo.Context) error {
		return w.controllers.Authenticate(c)
	})
	w.router.GET("/verify-email", func(c echo.Context) error {
		return w.controllers.VerifyEmail(c)
	})
	w.router.GET("/forgot-password", func(c echo.Context) error {
		return w.controllers.RenderPasswordForgotForm(c)
	})
	w.router.POST("/reset-password-request", func(c echo.Context) error {
		return w.controllers.SendPasswordResetEmail(c)
	})
	w.router.GET("/reset-password", func(c echo.Context) error {
		return w.controllers.ResetPasswordForm(c)
	})
	w.router.POST("/reset-password", func(c echo.Context) error {
		return w.controllers.ResetPassword(c)
	})
}
