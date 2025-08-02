package router

import (
	"encoding/hex"
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
	"github.com/mbvlabs/grafto/router/middleware"
	"github.com/mbvlabs/grafto/router/routes"
	"go.opentelemetry.io/otel/trace"
	"riverqueue.com/riverui"

	echomw "github.com/labstack/echo/v4/middleware"
)

type Router struct {
	e        *echo.Echo
	mw       middleware.MW
	handlers handlers.Handlers
}

func New(
	handlers handlers.Handlers,
	mw middleware.MW,
	riverUI *riverui.Server,
	traceProvider trace.TracerProvider,
) (*Router, error) {
	router := echo.New()
	router.Debug = true

	authKey, err := hex.DecodeString(config.Cfg.Auth.SessionKey)
	if err != nil {
		return nil, err
	}
	encKey, err := hex.DecodeString(config.Cfg.Auth.SessionEncryptionKey)
	if err != nil {
		return nil, err
	}

	router.Use(
		session.Middleware(
			sessions.NewCookieStore(
				authKey,
				encKey,
			),
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
				if config.Cfg.App.Env == config.PROD_ENVIRONMENT {
					return config.Cfg.App.Domain
				}

				return ""
			}(),
			CookieSecure:   config.Cfg.App.Env == config.PROD_ENVIRONMENT,
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

	return &Router{
		router,
		mw,
		handlers,
	}, nil
}

func (r *Router) SetupRoutes() *echo.Echo {
	setupRoutes(r.e, routes.AllRoutes, r.handlers, r.mw)
	r.setup404Handler()
	return r.e
}

func (r *Router) setup404Handler() {
	r.e.RouteNotFound("/*", getHandlerFunc(r.handlers.Pages, "NotFoundPage"))
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
	handlers handlers.Handlers,
	middlewares any,
) {
	registeredRoutes := []string{}
	handlersValue := reflect.ValueOf(handlers)

	for _, route := range r {
		if registered := slices.Contains(registeredRoutes, route.Name); registered {
			panic(
				fmt.Sprintf(
					"%s is registered more than once",
					route.Name,
				),
			)
		}

		if route.Handler == "" || route.HandleMethod == "" {
			panic("Route must specify Handler and HandleMethod fields")
		}

		handlerField := handlersValue.FieldByName(route.Handler)
		if !handlerField.IsValid() {
			panic(
				fmt.Sprintf(
					"Handler field %s not found in handlers struct",
					route.Handler,
				),
			)
		}

		handler := handlerField.Interface()
		handlerFunc := getHandlerFunc(handler, route.HandleMethod)
		middlewareFuncs := getAllMiddlewareFuncs(middlewares, route.Middleware)

		switch route.Method {
		case http.MethodGet:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.GET(route.Path, handlerFunc, middlewareFuncs...).Name = route.Name
		case http.MethodPost:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.POST(route.Path, handlerFunc, middlewareFuncs...).Name = route.Name
		case http.MethodPut:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.PUT(route.Path, handlerFunc, middlewareFuncs...).Name = route.Name
		case http.MethodDelete:
			registeredRoutes = append(registeredRoutes, route.Name)
			router.DELETE(route.Path, handlerFunc, middlewareFuncs...).Name = route.Name
		}
	}
}
