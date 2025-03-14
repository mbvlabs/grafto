package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func authRoutes(
	router *echo.Echo,
	handlers handlers.Authentication,
) {
	router.GET("/login", func(c echo.Context) error {
		return handlers.CreateAuthenticatedSession(c)
	}).Name = paths.CreateAuthenticatedSession.String()
	router.POST("/login", func(c echo.Context) error {
		return handlers.StoreAuthenticatedSession(c)
	}).Name = paths.StoreAuthenticatedSession.String()

	router.GET("/forgot-password", func(c echo.Context) error {
		return handlers.CreatePasswordReset(c)
	}).Name = paths.ForgotPassword.String()
	router.POST("/forgot-password", func(c echo.Context) error {
		return handlers.StorePasswordReset(c)
	}).Name = paths.StoreForgotPassword.String()

	router.GET("/reset-password", func(c echo.Context) error {
		return handlers.CreateResetPassword(c)
	}).Name = paths.ResetPassword.String()
	router.POST("/reset-password", func(c echo.Context) error {
		return handlers.StoreResetPassword(c)
	}).Name = paths.StoreForgotPassword.String()
}
