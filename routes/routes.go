package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/http/handlers"
	"github.com/mbv-labs/grafto/http/middleware"
)

type Routes struct {
	router               *echo.Echo
	appHandlers          handlers.App
	dashboardHandlers    handlers.Dashboard
	authHandlers         handlers.Authentication
	registrationHandlers handlers.Registration
	apiHandlers          handlers.Api
	baseHandlers         handlers.Base
	middleware           middleware.Middleware
	cfg                  config.Config
}

func NewRoutes(
	appHandlers handlers.App,
	dashboardHandlers handlers.Dashboard,
	authHandlers handlers.Authentication,
	registrationHandlers handlers.Registration,
	apiHandlers handlers.Api,
	baseHandlers handlers.Base,
	mw middleware.Middleware,
	cfg config.Config,
) *Routes {
	router := echo.New()

	if cfg.App.Environment == "development" {
		router.Debug = true
	}

	router.Static("/static", "static")
	router.Use(mw.RegisterUserContext)

	return &Routes{
		router,
		appHandlers,
		dashboardHandlers,
		authHandlers,
		registrationHandlers,
		apiHandlers,
		baseHandlers,
		mw,
		cfg,
	}
}

func (r *Routes) web() {
	authRoutes(r.router, r.authHandlers, r.middleware)
	errorRoutes(r.router, r.baseHandlers)
	dashboardRoutes(r.router, r.dashboardHandlers, r.middleware)
	appRoutes(r.router, r.appHandlers)
	registrationRoutes(r.router, r.registrationHandlers)
}

func (r *Routes) api() {
	apiV1Router := r.router.Group("/api/v1")
	apiV1Routes(apiV1Router, r.apiHandlers)
}

func (r *Routes) SetupRoutes() *echo.Echo {
	r.web()
	r.api()

	return r.router
}

func errorRoutes(router *echo.Echo, ctrl handlers.Base) {
	router.GET("/400", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})

	router.GET("/404", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})

	router.GET("/500", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})
}
