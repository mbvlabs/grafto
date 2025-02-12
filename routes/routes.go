package routes

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/http"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/static"
	"github.com/mbvlabs/grafto/views/paths"
	slogecho "github.com/samber/slog-echo"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	router   *echo.Echo
	handlers handlers.Handlers
}

func NewRoutes(
	handlers handlers.Handlers,
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
			echoprometheus.NewMiddleware(
				strings.Join(
					strings.Fields(strings.ToLower(config.Cfg.ProjectName)),
					"_",
				),
			),
		)
		router.GET("/metrics", echoprometheus.NewHandler())
	}

	echo.MustSubFS(static.Files, "static")
	router.StaticFS("/static", static.Files)

	router.Use(
		session.Middleware(
			sessions.NewCookieStore([]byte(config.Cfg.SessionEncryptionKey)),
		),
	)
	router.Use(http.RegisterAppContext)

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

	router.Any("/river*", echo.WrapHandler(riverUI), http.AuthOnly)

	return &Routes{
		router,
		handlers,
	}
}

func (r *Routes) web() {
	resourceRoutes(r.router)
	authRoutes(r.router, r.handlers.Authentication)
	dashboardRoutes(r.router, r.handlers.Dashboard)
	appRoutes(r.router, r.handlers.App)
	registrationRoutes(r.router, r.handlers.Registration)
}

func (r *Routes) api() {
	apiV1Router := r.router.Group("/api/v1")
	apiV1Routes(apiV1Router, r.handlers.Api)
}

func (r *Routes) SetupRoutes(
	ctx context.Context,
) (*echo.Echo, context.Context) {
	r.web()
	r.api()

	for _, route := range r.router.Routes() {
		ctx = context.WithValue(ctx, paths.Route(route.Name), route.Path)
	}

	return r.router, ctx
}
