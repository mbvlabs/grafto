package routes

import (
	"log/slog"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/controllers"
	"github.com/mbvlabs/grafto/server/middleware"
	slogecho "github.com/samber/slog-echo"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	router               *echo.Echo
	appHandlers          controllers.App
	dashboardHandlers    controllers.Dashboard
	authHandlers         controllers.Authentication
	registrationHandlers controllers.Registration
	apiHandlers          controllers.Api
	middleware           middleware.Middleware
}

func NewRoutes(
	appControllers controllers.App,
	dashboardControllers controllers.Dashboard,
	authControllers controllers.Authentication,
	registrationControllers controllers.Registration,
	apiControllers controllers.Api,
	mw middleware.Middleware,
	riverUI *riverui.Server,
) *Routes {
	router := echo.New()
	router.Debug = true

	if config.Cfg.Environment == config.PROD_ENVIRONMENT {
		router.Debug = false
		router.Use(echomw.GzipWithConfig(echomw.GzipConfig{
			Level: 5,
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Path(), "metrics")
			},
		}))
		router.Use(
			echoprometheus.NewMiddleware(config.Cfg.ProjectName),
		)
		router.GET("/metrics", echoprometheus.NewHandler())
	}
	router.Static("/static", "static")

	router.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Cfg.SessionEncryptionKey))))
	router.Use(mw.RegisterAppContext)

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

	router.Any("/river*", echo.WrapHandler(riverUI), mw.AuthOnly)

	return &Routes{
		router,
		appControllers,
		dashboardControllers,
		authControllers,
		registrationControllers,
		apiControllers,
		mw,
	}
}

func (r *Routes) web() {
	resourceRoutes(r.router)
	authRoutes(r.router, r.authHandlers)
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
