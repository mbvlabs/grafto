package middleware

import (
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
)

// AuthOnly is only a placeholder for now
func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
func CsrfMiddlewareFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		csrfToken := csrf.Token(c.Request())
		c.Response().Header().Set(echo.HeaderXCSRFToken, csrfToken)
		return next(c)
	}
}
