package handlers

import (
	"context"
	"encoding/gob"
	"io"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/handlers/middleware"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/contexts"
	"github.com/mbvlabs/grafto/services"
)

const (
	oneWeekInSeconds = 604800
)

type Handlers struct {
	Api            Api
	App            App
	Authentication Authentication
	Dashboard      Dashboard
	Registration   Registration
	Assets         Assets
	Fragments      Fragments
}

func setAppCtx(ctx echo.Context) context.Context {
	appcKey := contexts.AppKey{}
	appc := ctx.Get(appcKey.String())

	cOne := context.WithValue(
		ctx.Request().Context(),
		appcKey,
		appc,
	)

	flashCKey := contexts.FlashKey{}
	flashC := ctx.Get(flashCKey.String())

	return context.WithValue(
		cOne,
		flashCKey,
		flashC,
	)
}

//nolint:unused // needed helper method
func addFlash(
	c echo.Context, flashType contexts.FlashType, msg string,
) error {
	sess, err := session.Get(middleware.FlashSessionKey, c)
	if err != nil {
		return err
	}

	sess.AddFlash(contexts.FlashMessage{
		ID:        uuid.New(),
		Type:      flashType,
		CreatedAt: time.Now(),
		Message:   msg,
	}, middleware.FlashSessionKey)

	return sess.Save(c.Request(), c.Response())
}

func renderArgs(ctx echo.Context) (context.Context, io.Writer) {
	return setAppCtx(ctx), ctx.Response().Writer
}

func NewHandlers(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
	emailSvc services.EmailSender,
) Handlers {
	gob.Register(uuid.UUID{})
	gob.Register(contexts.FlashMessage{})

	api := newApi()
	app := newApp(db, cache)
	auth := newAuthentication(db, emailSvc)
	dashboard := newDashboard()
	registration := newRegistration(db, emailSvc)
	assets := newAssets()

	return Handlers{
		api,
		app,
		auth,
		dashboard,
		registration,
		assets,
		Fragments{},
	}
}

//nolint:unused // needed helper method
func redirectHx(w http.ResponseWriter, url string) error {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusSeeOther)

	return nil
}

//nolint:unused // needed helper method
func redirect(
	w http.ResponseWriter,
	r *http.Request,
	url string,
) error {
	http.Redirect(w, r, url, http.StatusSeeOther)
	return nil
}

func destroyAuthSession(
	c echo.Context,
) error {
	sess, err := session.Get(middleware.AuthenticatedSessionName, c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	sess.Values[middleware.SessIsAuthenticated] = false
	sess.Values[middleware.SessUserID] = ""
	sess.Values[middleware.SessUserEmail] = ""
	sess.Values[middleware.SessIsAdmin] = false

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

func createAuthSession(
	c echo.Context,
	extend bool,
	user models.UserEntity,
) error {
	sess, err := session.Get(middleware.AuthenticatedSessionName, c)
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
	sess.Values[middleware.SessIsAuthenticated] = true
	sess.Values[middleware.SessUserID] = user.ID
	sess.Values[middleware.SessUserEmail] = user.Email
	sess.Values[middleware.SessIsAdmin] = user.IsAdmin

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
