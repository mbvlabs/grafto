package handlers

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views/contexts"
)

var AuthenticatedSessionName = fmt.Sprintf(
	"ua-%s-%s",
	config.Cfg.ProjectName,
	config.Cfg.Environment,
)

const (
	SessIsAuthName   = "is_authenticated"
	SessUserID       = "user_id"
	SessUserEmail    = "user_email"
	SessIsAdmin      = "is_admin"
	oneWeekInSeconds = 604800
)

type Handlers struct {
	Api            Api
	App            App
	Authentication Authentication
	Dashboard      Dashboard
	Registration   Registration
}

func setAppCtx(ctx echo.Context) context.Context {
	appCtxKey := contexts.AppKey{}
	appCtx := ctx.Get(appCtxKey.String())

	return context.WithValue(
		ctx.Request().Context(),
		appCtxKey,
		appCtx,
	)
}

func renderArgs(ctx echo.Context) (context.Context, io.Writer) {
	return setAppCtx(ctx), ctx.Response().Writer
}

func NewHandlers(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, string],
	authSvc services.Auth,
	emailSvc services.Email,
) Handlers {
	gob.Register(uuid.UUID{})

	api := newApi()
	app := newApp(db, cache)
	auth := newAuthentication(authSvc, db, emailSvc)
	dashboard := newDashboard()
	registration := newRegistration(authSvc, db, emailSvc)

	return Handlers{
		api,
		app,
		auth,
		dashboard,
		registration,
	}
}

func redirectHx(w http.ResponseWriter, url string) error {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusSeeOther)

	return nil
}

func getContext(c echo.Context) context.Context {
	return c.Request().Context()
}

func redirect(
	w http.ResponseWriter,
	r *http.Request,
	url string,
) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func createAuthSession(
	c echo.Context,
	extend bool,
	user models.UserEntity,
) error {
	sess, err := session.Get(AuthenticatedSessionName, c)
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
	sess.Values[SessIsAuthName] = true
	sess.Values[SessUserID] = user.ID
	sess.Values[SessUserEmail] = user.Email
	sess.Values[SessIsAdmin] = user.IsAdmin

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
