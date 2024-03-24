package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/services"
)

type Middleware struct {
	authSessionStore *sessions.CookieStore
}

func NewMiddleware(aSS *sessions.CookieStore) Middleware {
	return Middleware{aSS}
}

type UserContext struct {
	echo.Context
	UserID          uuid.UUID
	IsAuthenticated bool
}

func (u *UserContext) GetID() uuid.UUID {
	return u.UserID
}

func (u *UserContext) GetAuthStatus() bool {
	return u.IsAuthenticated
}

type AdminContext struct {
	echo.Context
	isAdmin bool
}

func (a *AdminContext) GetAdminStatus() bool {
	return a.isAdmin
}

func (m *Middleware) AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authenticated, userID, err := services.IsAuthenticated(c.Request(), m.authSessionStore)
		if err != nil {
			telemetry.Logger.Error("could not get authenticated status", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		if authenticated {
			ctx := &UserContext{c, userID, true}
			return next(ctx)
		} else {
			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}
	}
}

func (m *Middleware) RegisterUserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authenticated, userID, err := services.IsAuthenticated(c.Request(), m.authSessionStore)
		if err != nil {
			telemetry.Logger.Error("could not get authenticated status", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		authContext := &UserContext{
			c,
			userID,
			authenticated,
		}

		return next(authContext)
	}
}

func (m *Middleware) AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin, err := services.IsAdmin(c.Request(), m.authSessionStore)
		if err != nil {
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		ctx := &AdminContext{c, isAdmin}
		return next(ctx)
	}
}
