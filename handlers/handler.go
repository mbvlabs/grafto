package handlers

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/routes/contexts"
)

var AuthenticatedSessionName = fmt.Sprintf(
	"ua-%s-%s",
	strings.ToLower(config.Cfg.ProjectName),
	config.Cfg.Environment,
)

const (
	FlashSessionKey     = "flash_messages"
	SessIsAuthenticated = "is_authenticated"
	SessUserID          = "user_id"
	SessUserEmail       = "user_email"
	SessIsAdmin         = "is_admin"
	oneWeekInSeconds    = 604800
)

type Handlers struct {
	Api            Api
	App            App
	Authentication Authentication
	Dashboard      Dashboard
	Registration   Registration
	Assets         Assets
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
	sess, err := session.Get(FlashSessionKey, c)
	if err != nil {
		return err
	}

	sess.AddFlash(contexts.FlashMessage{
		ID:        uuid.New(),
		Type:      flashType,
		CreatedAt: time.Now(),
		Message:   msg,
	}, FlashSessionKey)

	return sess.Save(c.Request(), c.Response())
}

func renderArgs(ctx echo.Context) (context.Context, io.Writer) {
	return setAppCtx(ctx), ctx.Response().Writer
}

type EmailClient interface {
	Send(
		ctx context.Context,
		payload clients.EmailPayload,
	) error
}

func NewHandlers(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
	emailSvc EmailClient,
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
	sess, err := session.Get(AuthenticatedSessionName, c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	sess.Values[SessIsAuthenticated] = false
	sess.Values[SessUserID] = ""
	sess.Values[SessUserEmail] = ""
	sess.Values[SessIsAdmin] = false

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
	sess.Values[SessIsAuthenticated] = true
	sess.Values[SessUserID] = user.ID
	sess.Values[SessUserEmail] = user.Email
	sess.Values[SessIsAdmin] = user.IsAdmin

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
