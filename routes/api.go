package routes

import (
	"time"

	"github.com/mbv-labs/grafto/controllers"
	"github.com/mbv-labs/grafto/server/middleware"
	"github.com/labstack/echo/v4"
)

func apiRoutes(router *echo.Group, controllers controllers.Controller, middleware middleware.Middleware) {
	router.GET("/health", func(c echo.Context) error {
		time.Sleep(6 * time.Second)
		return controllers.AppHealth(c)
	})

}
