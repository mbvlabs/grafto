package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
)

func authRoutes(
	router *echo.Echo,
	handlers handlers.Authentication,
) {
	router.GET("/login", func(c echo.Context) error {
		return handlers.CreateAuthenticatedSession(c)
	})
	router.POST("/login", func(c echo.Context) error {
		return handlers.StoreAuthenticatedSession(c)
	})

	router.GET("/forgot-password", func(c echo.Context) error {
		return handlers.CreatePasswordReset(c)
	})
	router.POST("/forgot-password", func(c echo.Context) error {
		return handlers.StorePasswordReset(c)
	})
	router.GET("/reset-password", func(c echo.Context) error {
		return handlers.CreateResetPassword(c)
	})
	router.POST("/reset-password", func(c echo.Context) error {
		return handlers.StoreResetPassword(c)
	})
}
