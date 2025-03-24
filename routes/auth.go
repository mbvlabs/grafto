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
	router.GET(paths.CreateAuthenticatedSession.URL, func(c echo.Context) error {
		return handlers.CreateAuthenticatedSession(c)
	}).Name = paths.CreateAuthenticatedSession.Name
	router.POST(paths.StoreAuthenticatedSession.URL, func(c echo.Context) error {
		return handlers.StoreAuthenticatedSession(c)
	}).Name = paths.StoreAuthenticatedSession.Name

	router.GET(paths.CreateForgotPassword.URL, func(c echo.Context) error {
		return handlers.CreatePasswordReset(c)
	}).Name = paths.CreateForgotPassword.Name
	router.POST(paths.StoreForgotPassword.URL, func(c echo.Context) error {
		return handlers.StorePasswordReset(c)
	}).Name = paths.StoreForgotPassword.Name

	router.GET(paths.CreateResetPassword.URL, func(c echo.Context) error {
		return handlers.CreateResetPassword(c)
	}).Name = paths.CreateResetPassword.Name
	router.POST(paths.StoreResetPassword.URL, func(c echo.Context) error {
		return handlers.StoreResetPassword(c)
	}).Name = paths.StoreResetPassword.Name
}
