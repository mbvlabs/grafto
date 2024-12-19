package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/server"
)

const AppContextName = "APP_CONTEXT"

type Middleware struct{}

func NewMiddleware() Middleware {
	return Middleware{}
}

func (m *Middleware) AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return next(c)
		}

		isAuth, _ := sess.Values[server.SessIsAuthName].(bool)
		if isAuth {
			return next(c)
		}

		return c.Redirect(http.StatusPermanentRedirect, "/login")
	}
}

func (m *Middleware) RegisterAppContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(server.AuthenticatedSessionName, c)
		if err != nil {
			return err
		}

		isAuth, _ := sess.Values[server.SessIsAuthName].(bool)
		userID, _ := sess.Values[server.SessUserID].(uuid.UUID)
		userEmail, _ := sess.Values[server.SessUserEmail].(string)
		isAdmin, _ := sess.Values[server.SessIsAdmin].(bool)

		c.Set(AppContextName, &AppContext{
			c,
			userID,
			userEmail,
			isAuth,
			isAdmin,
			c.Request().URL.Path,
		})

		return next(c)
	}
}
