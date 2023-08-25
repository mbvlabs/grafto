package api

import (
	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/labstack/echo/v4"
)

type API struct {
	controllers controllers.Controller
	router      *echo.Group
}

func NewAPI(router *echo.Echo, controllers controllers.Controller) API {
	apiRouter := router.Group("/api")
	return API{
		controllers: controllers,
		router:      apiRouter,
	}
}

func (a *API) SetupAPIRoutes() {
	a.router.GET("/health", func(c echo.Context) error {
		return a.controllers.AppHealth(c)
	}, middleware.AuthOnly)
}
