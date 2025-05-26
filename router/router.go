package router

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/handlers/middleware"
	"github.com/mbvlabs/grafto/router/routes"
	"go.opentelemetry.io/otel/trace"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	router   *echo.Echo
	mw       middleware.MW
	handlers handlers.Handlers
}

func New(
	ctx context.Context,
	handlers handlers.Handlers,
	mw middleware.MW,
	riverUI *riverui.Server,
	traceProvider trace.TracerProvider,
) *Routes {
	router := echo.New()
	router.Debug = true

	router.Use(
		session.Middleware(
			sessions.NewCookieStore([]byte(config.Cfg.SessionEncryptionKey)),
		),
		mw.RegisterAppContext,
		mw.RegisterFlashMessagesContext,

		echomw.CSRFWithConfig(echomw.CSRFConfig{
			Skipper: func(c echo.Context) bool {
				if strings.HasPrefix(c.Request().URL.Path, "/api") ||
					strings.HasPrefix(c.Request().URL.Path, "/river") {
					return true
				}

				return false
			},
			TokenLookup: "cookie:_csrf",
			CookiePath:  "/",
			CookieDomain: func() string {
				if config.Cfg.Environment == config.PROD_ENVIRONMENT {
					return config.Cfg.GetFullDomain()
				}

				return ""
			}(),
			CookieSecure:   config.Cfg.Environment == config.PROD_ENVIRONMENT,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteStrictMode,
		}),
	)

	router.Use(
		//nolint:contextcheck // not needed here
		mw.Logging(),
		echomw.Recover(),
	)

	router.Any("/river*", echo.WrapHandler(riverUI), mw.AuthOnly)

	return &Routes{
		router,
		mw,
		handlers,
	}
}

func (r *Routes) SetupRoutes(
	ctx context.Context,
) (*echo.Echo, context.Context) {
	setupRoutes(r.router, routes.Assets, r.handlers.Assets, r.mw)
	setupRoutes(
		r.router,
		routes.Authentication,
		r.handlers.Authentication,
		r.mw,
	)
	setupRoutes(r.router, routes.Dashboard, r.handlers.Dashboard, r.mw)
	setupRoutes(r.router, routes.App, r.handlers.App, r.mw)
	setupRoutes(
		r.router,
		routes.Registration,
		r.handlers.Registration,
		r.mw,
	)
	setupRoutes(r.router, routes.ApiV1, r.handlers.Api, r.mw)

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

func getAllMiddlewareFuncs(
	middlewares any,
	middlewareNames []string,
) []echo.MiddlewareFunc {
	var middlewareFuncs []echo.MiddlewareFunc

	for _, name := range middlewareNames {
		middlewareFuncs = append(
			middlewareFuncs,
			getMiddlewareFunc(middlewares, name),
		)
	}

	return middlewareFuncs
}

func getMiddlewareFunc(handlers any, methodName string) echo.MiddlewareFunc {
	appValue := reflect.ValueOf(handlers)
	appType := appValue.Type()

	method, found := appType.MethodByName(methodName)
	if !found {
		panic(fmt.Sprintf("Handler method %s not found", methodName))
	}

	methodType := method.Type
	numIn := methodType.NumIn()
	numOut := methodType.NumOut()

	if numOut != 1 {
		panic(
			fmt.Sprintf("Method %s must return exactly one value", methodName),
		)
	}

	returnType := methodType.Out(0)
	handlerFuncType := reflect.TypeOf((echo.HandlerFunc)(nil))
	middlewareFuncType := reflect.TypeOf((echo.MiddlewareFunc)(nil))

	switch numIn {
	case 1:
		// Signature: func() echo.MiddlewareFunc
		if !returnType.AssignableTo(middlewareFuncType) {
			panic(
				fmt.Sprintf(
					"Method %s must return echo.MiddlewareFunc",
					methodName,
				),
			)
		}
		values := method.Func.Call([]reflect.Value{appValue})
		middleware, _ := values[0].Interface().(echo.MiddlewareFunc)
		if middleware == nil {
			panic(fmt.Sprintf("Method %s returned nil", methodName))
		}
		return middleware

	case 2:
		// Signature: func(echo.HandlerFunc) echo.HandlerFunc
		if !returnType.AssignableTo(handlerFuncType) {
			panic(
				fmt.Sprintf(
					"Method %s must return echo.HandlerFunc",
					methodName,
				),
			)
		}
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			values := method.Func.Call([]reflect.Value{
				appValue,
				reflect.ValueOf(next),
			})
			return values[0].Interface().(echo.HandlerFunc)
		}

	default:
		panic(
			fmt.Sprintf(
				"Method %s has unsupported number of parameters",
				methodName,
			),
		)
	}
}

func setupRoutes(
	router *echo.Echo,
	r []routes.Route,
	handlers any,
	middlewares any,
) {
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
			router.GET(route.Path, getHandlerFunc(handlers, route.HandlerName), getAllMiddlewareFuncs(middlewares, route.Middleware)...).Name = route.Name
		case http.MethodPost:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.POST(route.Path, getHandlerFunc(handlers, route.HandlerName), getAllMiddlewareFuncs(middlewares, route.Middleware)...).Name = route.Name
		case http.MethodPut:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.PUT(route.Path, getHandlerFunc(handlers, route.HandlerName), getAllMiddlewareFuncs(middlewares, route.Middleware)...).Name = route.Name
		case http.MethodDelete:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.DELETE(route.Path, getHandlerFunc(handlers, route.HandlerName), getAllMiddlewareFuncs(middlewares, route.Middleware)...).Name = route.Name
		}
	}
}
