package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/services"
)

type Middleware struct {
	authSvc services.Auth
}

func NewMiddleware(authSvc services.Auth) Middleware {
	return Middleware{authSvc}
}

func (m *Middleware) AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := m.authSvc.GetUserSession(c.Request())
		if err != nil {
			slog.ErrorContext(
				c.Request().Context(),
				"could not get user session",
				"error",
				err,
			)

			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}

		if sess.Authenticated {
			ctx := &UserContext{c, sess.ID, true}
			return next(ctx)
		} else {
			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}
	}
}

func (m *Middleware) RegisterUserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := m.authSvc.GetUserSession(c.Request())
		if err != nil {
			telemetry.Logger.Error("could not get authenticated user session", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		authContext := &UserContext{
			c,
			sess.ID,
			sess.Authenticated,
		}

		return next(authContext)
	}
}

func (m *Middleware) AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := m.authSvc.GetUserSession(c.Request())
		if err != nil {
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		ctx := &AdminContext{c, sess.IsAdmin}
		return next(ctx)
	}
}
