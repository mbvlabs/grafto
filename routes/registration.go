package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/http/handlers"
)

func registrationRoutes(
	router *echo.Echo,
	controllers handlers.Registration,
) {
	router.GET("/register", func(c echo.Context) error {
		return controllers.CreateUser(c)
	})
	router.POST("/register", func(c echo.Context) error {
		return controllers.StoreUser(c)
	})

	router.GET("/verify-email", func(c echo.Context) error {
		return controllers.VerifyUserEmail(c)
	})
}
