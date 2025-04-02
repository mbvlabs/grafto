package middleware

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/handlers"
)

func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(handlers.AuthenticatedSessionName, c)
		if err != nil {
			return next(c)
		}

		isAuth, _ := sess.Values[handlers.SessIsAuthenticated].(bool)
		if isAuth {
			return next(c)
		}

		return c.Redirect(http.StatusPermanentRedirect, "/login")
	}
}
