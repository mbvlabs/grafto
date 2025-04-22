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
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/routes/middleware"
	"github.com/mbvlabs/grafto/routes/paths"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	router   *echo.Echo
	handlers handlers.Handlers
}

func NewRoutes(
	ctx context.Context,
	handlers handlers.Handlers,
	riverUI *riverui.Server,
) *Routes {
	router := echo.New()
	router.Debug = true

	router.Use(
		session.Middleware(
			sessions.NewCookieStore([]byte(config.Cfg.SessionEncryptionKey)),
		),
		middleware.RegisterAppContext,
		middleware.RegisterFlashMessagesContext,
	)

	if config.Cfg.Environment == config.PROD_ENVIRONMENT {
		router.Debug = false
		router.Use(
			echomw.GzipWithConfig(echomw.GzipConfig{
				Level: 5,
				Skipper: func(c echo.Context) bool {
					return strings.Contains(c.Path(), "metrics")
				},
			}),

			echoprometheus.NewMiddleware(
				strings.Join(
					strings.Fields(strings.ToLower(config.Cfg.ProjectName)),
					"_",
				),
			),
		)

		router.GET("/metrics", echoprometheus.NewHandler())
	}

	// slogechoCfg := slogecho.Config{
	// 	WithRequestID: false,
	// 	WithTraceID:   false,
	// 	Filters: []slogecho.Filter{
	// 		slogecho.IgnorePathContains("static"),
	// 		slogecho.IgnorePathContains("health"),
	// 	},
	// }

	router.Use(
		// slogecho.NewWithConfig(slog.Default(), slogechoCfg),
		setupLogger(ctx),
		echomw.Recover(),
	)

	router.Any("/river*", echo.WrapHandler(riverUI), middleware.AuthOnly)

	return &Routes{
		router,
		handlers,
	}
}

func setupLogger(ctx context.Context) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogStatus:   true,
		LogHost:     true,
		LogMethod:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/assets")
		},
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			level := slog.LevelInfo
			attrs := []slog.Attr{
				slog.String(
					"timestamp",
					v.StartTime.Format("2006-01-02 15:04:05 MST -0700"),
				),
				slog.Int("status", v.Status),
				slog.String("uri", v.URI),
				slog.String("method", v.Method),
				slog.String("host", v.Host),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				level = slog.LevelError
			}

			slog.Default().LogAttrs(ctx, level, "req",
				attrs...,
			)

			return nil
		},
	})
}

func (r *Routes) web() {
	assetsRoutes(r.router, r.handlers.Assets)
	fragmentRoutes(r.router)
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
		ctx = context.WithValue(
			ctx, paths.RouteCtxKey(route.Name), route.Path,
		)
	}

	return r.router, ctx
}
