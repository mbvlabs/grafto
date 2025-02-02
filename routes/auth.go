package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/views/paths"
)

func authRoutes(
	router *echo.Echo,
	handlers handlers.Authentication,
) {
	router.GET("/login", func(c echo.Context) error {
		return handlers.CreateAuthenticatedSession(c)
	}).Name = paths.LoginPage.Name()
	router.POST("/login", func(c echo.Context) error {
		return handlers.StoreAuthenticatedSession(c)
	}).Name = paths.Login.Name()

	router.GET("/forgot-password", func(c echo.Context) error {
		return handlers.CreatePasswordReset(c)
	}).Name = paths.ForgotPasswordPage.Name()
	router.POST("/forgot-password", func(c echo.Context) error {
		return handlers.StorePasswordReset(c)
	}).Name = paths.ForgotPassword.Name()

	router.GET("/reset-password", func(c echo.Context) error {
		return handlers.CreateResetPassword(c)
	}).Name = paths.ResetPasswordPage.Name()
	router.POST("/reset-password", func(c echo.Context) error {
		return handlers.StoreResetPassword(c)
	}).Name = paths.ResetPassword.Name()
}
