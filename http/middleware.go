package http

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/contexts"
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

func RegisterAppContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, "/static") ||
			strings.HasPrefix(c.Request().URL.Path, "/fragments") {
			return next(c)
		}

		sess, err := session.Get(handlers.AuthenticatedSessionName, c)
		if err != nil {
			return err
		}

		isAuth, _ := sess.Values[handlers.SessIsAuthenticated].(bool)
		userID, _ := sess.Values[handlers.SessUserID].(uuid.UUID)
		userEmail, _ := sess.Values[handlers.SessUserEmail].(string)
		isAdmin, _ := sess.Values[handlers.SessIsAdmin].(bool)

		ac := contexts.App{
			Context:         c,
			UserID:          userID,
			Email:           userEmail,
			IsAuthenticated: isAuth,
			IsAdmin:         isAdmin,
			CurrentPath:     c.Request().URL.Path,
		}

		c.Set(contexts.AppKey{}.String(), ac)

		return next(c)
	}
}

func RegisterFlashMessagesContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, "/static") {
			return next(c)
		}

		sess, err := session.Get(handlers.FlashSessionKey, c)
		if err != nil {
			return err
		}

		flashMessages := []contexts.FlashMessage{}
		if flashes := sess.Flashes(handlers.FlashSessionKey); len(
			flashes,
		) > 0 {
			for _, flash := range flashes {
				if msg, ok := flash.(contexts.FlashMessage); ok {
					flashMessages = append(
						flashMessages,
						contexts.FlashMessage{
							Context:   c,
							ID:        msg.ID,
							Type:      msg.Type,
							CreatedAt: msg.CreatedAt,
							Message:   msg.Message,
						},
					)
				}
			}

			if err := sess.Save(c.Request(), c.Response()); err != nil {
				return err
			}
		}

		c.Set(contexts.FlashKey{}.String(), flashMessages)

		return next(c)
	}
}
