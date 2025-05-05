package routes

import "github.com/labstack/echo/v4"

// type Cache struct {
// 	MaxAge string
// 	Etag   string
// }

type Route struct {
	Name       string
	Path       string
	CtrlName   string
	Method     string
	Middleware []echo.MiddlewareFunc
	// Caching    Cache
}
