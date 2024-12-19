package controllers

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/server"
)

const (
	oneWeekInSeconds = 604800
)

// type Base struct {
// 	cfg         config.Config
// 	db          psql.Postgres
// 	flashStore  FlashStorage
// 	queueClient *river.Client[pgx.Tx]
// 	tracer      telemetry.Tracer
// }
//
// func NewDependencies(
// 	cfg config.Config,
// 	db psql.Postgres,
// 	flashStore FlashStorage,
// 	queueClient *river.Client[pgx.Tx],
// 	tracer telemetry.Tracer,
// ) Base {
// 	return Base{
// 		cfg,
// 		db,
// 		flashStore,
// 		queueClient,
// 		tracer,
// 	}
// }

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

func internalError(ctx echo.Context) error {
	return ctx.HTML(
		http.StatusOK,
		"<h2>An unrecoverable error occurred. Please click <a href='/'>here</a></h2>",
	)
}

func createAuthSession(c echo.Context, extend bool, user models.UserEntity) error {
	sess, err := session.Get(server.AuthenticatedSessionName, c)
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
	sess.Values[server.SessIsAuthName] = true
	sess.Values[server.SessUserID] = user.ID
	sess.Values[server.SessUserEmail] = user.Email
	sess.Values[server.SessIsAdmin] = user.IsAdmin

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
