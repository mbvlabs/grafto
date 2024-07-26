package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/http/handlers"
	"github.com/mbv-labs/grafto/http/middleware"
)

func authRoutes(
	router *echo.Echo,
	controllers handlers.Authentication,
	middleware middleware.Middleware,
) {
	router.GET("/login", func(c echo.Context) error {
		return controllers.CreateAuthenticatedSession(c)
	})
	router.POST("/login", func(c echo.Context) error {
		return controllers.StoreAuthenticatedSession(c)
	})

	router.GET("/forgot-password", func(c echo.Context) error {
		return controllers.CreatePasswordReset(c)
	})
	router.POST("/forgot-password", func(c echo.Context) error {
		return controllers.StorePasswordReset(c)
	})
	router.GET("/reset-password", func(c echo.Context) error {
		return controllers.CreateResetPassword(c)
	})
	router.POST("/reset-password", func(c echo.Context) error {
		return controllers.StoreResetPassword(c)
	})
}
