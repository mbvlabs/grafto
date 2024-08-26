package routes

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/http/middleware"
	slogecho "github.com/samber/slog-echo"

	echomw "github.com/labstack/echo/v4/middleware"
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

	router.Debug = true

	if cfg.Environment == config.PROD_ENVIRONMENT {
		router.Debug = false
		router.Use(echomw.GzipWithConfig(echomw.GzipConfig{
			Level: 5,
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Path(), "metrics")
			},
		}))
		router.Use(
			echoprometheus.NewMiddleware(cfg.ProjectName),
		)
		router.GET("/metrics", echoprometheus.NewHandler())
	}

	router.Static("/static", "static")
	router.Use(mw.RegisterUserContext)

	slogechoCfg := slogecho.Config{
		WithRequestID: false,
		WithTraceID:   false,
		Filters: []slogecho.Filter{
			slogecho.IgnorePathContains("static"),
			slogecho.IgnorePathContains("health"),
		},
	}
	router.Use(slogecho.NewWithConfig(slog.Default(), slogechoCfg))

	router.Use(echomw.Recover())

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
