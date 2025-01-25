package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/views"
)

func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return next(c)
		}

		isAuth, _ := sess.Values[handlers.SessIsAuthName].(bool)
		if isAuth {
			return next(c)
		}

		return c.Redirect(http.StatusPermanentRedirect, "/login")
	}
}

func RegisterAppContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(handlers.AuthenticatedSessionName, c)
		if err != nil {
			return err
		}

		isAuth, _ := sess.Values[handlers.SessIsAuthName].(bool)
		userID, _ := sess.Values[handlers.SessUserID].(uuid.UUID)
		userEmail, _ := sess.Values[handlers.SessUserEmail].(string)
		isAdmin, _ := sess.Values[handlers.SessIsAdmin].(bool)

		ac := &views.AppContext{
			Context:         c,
			UserID:          userID,
			Email:           userEmail,
			IsAuthenticated: isAuth,
			IsAdmin:         isAdmin,
			CurrentPath:     c.Request().URL.Path,
		}

		c.Set(views.AppContextKey{}.Value(), ac)

		return next(c)
	}
}
