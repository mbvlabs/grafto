package routes

import (
	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/server/middleware"
	"github.com/labstack/echo/v4"
)

type Routes struct {
	router      *echo.Echo
	controllers controllers.Controller
	middleware  middleware.Middleware
	cfg         config.Cfg
}

func NewRoutes(ctrl controllers.Controller, mw middleware.Middleware, cfg config.Cfg) *Routes {
	router := echo.New()

	if cfg.App.Environment == "development" {
		router.Debug = true
	}

	router.Static("/static", "static")

	return &Routes{
		router:      router,
		controllers: ctrl,
		middleware:  mw,
	}
}

func (r *Routes) web() {
	authRoutes(r.router, r.controllers, r.middleware)
	errorRoutes(r.router, r.controllers, r.middleware)
	appRoutes(r.router, r.controllers)
}

func (r *Routes) api() {
	apiRouter := r.router.Group("/api")
	apiRoutes(apiRouter, r.controllers, r.middleware)
}

func (r *Routes) SetupRoutes() *echo.Echo {
	r.web()
	r.api()

	return r.router
}

func appRoutes(router *echo.Echo, ctrl controllers.Controller) {
	router.GET("/", func(c echo.Context) error {
		return ctrl.LandingPage(c)
	})
}

func errorRoutes(router *echo.Echo, ctrl controllers.Controller, mw middleware.Middleware) {
	router.GET("/400", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})

	router.GET("/404", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})

	router.GET("/500", func(c echo.Context) error {
		return ctrl.InternalError(c)
	})

	router.GET("/redirect", func(c echo.Context) error {
		return ctrl.Redirect(c)
	})
}
