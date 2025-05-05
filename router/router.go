package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/router/middleware"
	"github.com/mbvlabs/grafto/router/routes"
	slogecho "github.com/samber/slog-echo"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	router   *echo.Echo
	handlers handlers.Handlers
}

func New(
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

	slogechoCfg := slogecho.Config{
		WithRequestID: false,
		WithTraceID:   false,
		Filters: []slogecho.Filter{
			slogecho.IgnorePathContains("static"),
			slogecho.IgnorePathContains("health"),
		},
	}

	router.Use(
		slogecho.NewWithConfig(slog.Default(), slogechoCfg),
		echomw.Recover(),
	)

	router.Any("/river*", echo.WrapHandler(riverUI), middleware.AuthOnly)

	return &Routes{
		router,
		handlers,
	}
}

func (r *Routes) SetupRoutes(
	ctx context.Context,
) (*echo.Echo, context.Context) {
	setupRoutes(r.router, routes.Assets, r.handlers.Assets)
	setupRoutes(r.router, routes.Authentication, r.handlers.Authentication)
	setupRoutes(r.router, routes.Dashboard, r.handlers.Dashboard)
	setupRoutes(r.router, routes.App, r.handlers.App)
	setupRoutes(r.router, routes.Registration, r.handlers.Registration)
	setupRoutes(r.router, routes.ApiV1, r.handlers.Api)

	return r.router, ctx
}

func getHandlerFunc(handlers any, methodName string) echo.HandlerFunc {
	appType := reflect.TypeOf(handlers)
	method, found := appType.MethodByName(methodName)
	if !found {
		panic(fmt.Sprintf("Handler method %s not found", methodName))
	}

	return func(c echo.Context) error {
		values := method.Func.Call([]reflect.Value{
			reflect.ValueOf(handlers),
			reflect.ValueOf(c),
		})

		if len(values) != 1 {
			panic(
				fmt.Sprintf(
					"Handler %s does not return exactly one value",
					methodName,
				),
			)
		}

		if values[0].IsNil() {
			return nil
		}

		return values[0].Interface().(error)
	}
}

func setupRoutes(router *echo.Echo, r []routes.Route, handlers any) {
	registeredRoutes := []string{}
	for _, route := range r {
		if registered := slices.Contains(registeredRoutes, route.Name); registered {
			panic(
				fmt.Sprintf(
					"%s is registered more than once",
					route.Name,
				),
			)
		}
		switch route.Method {
		case http.MethodGet:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.GET(route.Path, func(c echo.Context) error {
				return getHandlerFunc(handlers, route.CtrlName)(c)
			}).Name = route.Name
		case http.MethodPost:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.POST(route.Path, getHandlerFunc(handlers, route.CtrlName)).Name = route.Name
		case http.MethodPut:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.PUT(route.Path, getHandlerFunc(handlers, route.CtrlName)).Name = route.Name
		case http.MethodDelete:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.DELETE(route.Path, getHandlerFunc(handlers, route.CtrlName)).Name = route.Name
		}
	}
}
