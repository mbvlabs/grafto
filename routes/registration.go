package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
)

func registrationRoutes(
	router *echo.Echo,
	handlers handlers.Registration,
) {
	router.GET("/register", func(c echo.Context) error {
		return handlers.CreateUser(c)
	})
	router.POST("/register", func(c echo.Context) error {
		return handlers.StoreUser(c)
	})

	router.GET("/verify-email", func(c echo.Context) error {
		return handlers.VerifyUserEmail(c)
	})
}
