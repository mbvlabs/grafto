package routes

import (
	"github.com/mbv-labs/grafto/controllers"
	"github.com/mbv-labs/grafto/server/middleware"
	"github.com/labstack/echo/v4"
)

func authRoutes(router *echo.Echo, controllers controllers.Controller, middleware middleware.Middleware) {
	router.GET("/register", func(c echo.Context) error {
		return controllers.CreateUser(c)
	})
	router.POST("/register", func(c echo.Context) error {
		return controllers.StoreUser(c)
	})

	router.GET("/login", func(c echo.Context) error {
		return controllers.CreateAuthenticatedSession(c)
	})
	router.POST("/login", func(c echo.Context) error {
		return controllers.StoreAuthenticatedSession(c)
	})

	router.GET("/verify-email", func(c echo.Context) error {
		return controllers.VerifyEmail(c)
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
