package api

import (
	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/labstack/echo/v4"
)

type API struct {
	controllers controllers.Controller
	router      *echo.Group
	logger      *slog.Logger
}

func NewAPI(router *echo.Echo, controllers controllers.Controller, logger *slog.Logger) API {
	apiRouter := router.Group("/api")
	return API{
		controllers: controllers,
		logger:      logger,
		router:      apiRouter,
	}
}

func (a *API) SetupAPIRoutes() {
	a.router.GET("/health", func(c echo.Context) error {
		return a.controllers.AppHealth(c)
	}, middleware.AuthOnly)
}
