package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/views/paths"
)

func registrationRoutes(
	router *echo.Echo,
	handlers handlers.Registration,
) {
	router.GET("/register", func(c echo.Context) error {
		return handlers.CreateUser(c)
	}).Name = paths.RegisterPage.Name()
	router.POST("/register", func(c echo.Context) error {
		return handlers.StoreUser(c)
	}).Name = paths.RegisterUser.Name()

	router.GET("/verify-email", func(c echo.Context) error {
		return handlers.VerifyUserEmail(c)
	}).Name = paths.VerifyEmailPage.Name()
}
