package cookies

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
)

func CreateAuth(
	c echo.Context,
	extend bool,
	user models.User,
) error {
	sess, err := session.Get(authenticatedSessionName, c)
	if err != nil {
		return err
	}

	maxAge := oneWeekInSeconds
	if extend {
		maxAge = oneWeekInSeconds * 2
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
	}
	sess.Values[isAuthenticated] = true
	sess.Values[userID] = user.ID
	sess.Values[userEmail] = user.Email
	sess.Values[userEmail] = user.IsAdmin

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

func DestroyAuthSession(
	c echo.Context,
) error {
	sess, err := session.Get(authenticatedSessionName, c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	sess.Values[isAuthenticated] = false
	sess.Values[userID] = ""
	sess.Values[userEmail] = ""
	sess.Values[userEmail] = false

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
