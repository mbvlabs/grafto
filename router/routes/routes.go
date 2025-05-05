package routes

import "github.com/labstack/echo/v4"

type Route struct {
	Name       string
	Path       string
	CtrlName   string
	Method     string
	Middleware []echo.MiddlewareFunc
}
