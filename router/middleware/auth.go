package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/router/cookies"
)

func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if cookies.GetApp(c).IsAdmin {
			return next(c)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/sessions/new")
	}
}
