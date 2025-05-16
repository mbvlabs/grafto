package middleware

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (m MW) AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(AuthenticatedSessionName, c)
		if err != nil {
			return next(c)
		}

		isAuth, _ := sess.Values[SessIsAuthenticated].(bool)
		if isAuth {
			return next(c)
		}

		return c.Redirect(http.StatusPermanentRedirect, "/login")
	}
}
