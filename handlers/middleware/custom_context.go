package middleware

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/router/contexts"
)

func (m MW) RegisterAppContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, "/static") ||
			strings.HasPrefix(c.Request().URL.Path, "/fragments") {
			return next(c)
		}

		sess, err := session.Get(AuthenticatedSessionName, c)
		if err != nil {
			return err
		}

		isAuth, _ := sess.Values[SessIsAuthenticated].(bool)
		userID, _ := sess.Values[SessUserID].(uuid.UUID)
		userEmail, _ := sess.Values[SessUserEmail].(string)
		isAdmin, _ := sess.Values[SessIsAdmin].(bool)

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

func (m MW) RegisterFlashMessagesContext(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Request().URL.Path, "/static") {
			return next(c)
		}

		sess, err := session.Get(FlashSessionKey, c)
		if err != nil {
			return err
		}

		flashMessages := []contexts.FlashMessage{}
		if flashes := sess.Flashes(FlashSessionKey); len(
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
