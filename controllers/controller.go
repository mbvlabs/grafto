package controllers

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
)

const (
	oneWeekInSeconds = 604800
)

type Controllers struct {
	API           Api
	Pages         Pages
	Sessions      Sessions
	Dashboard     Dashboard
	Registrations Registrations
	Assets        Assets
	Fragments     Fragments
}

func renderArgs(ctx echo.Context) (context.Context, io.Writer) {
	return ctx.Request().Context(), ctx.Response().Writer
}

func New(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
	emailSvc services.EmailSender,
) Controllers {
	api := newApi()
	pages := newPages(db, cache)
	auth := newSessions(db, emailSvc)
	dashboard := newDashboard(db)
	registration := newRegistrations(db, emailSvc)
	assets := newAssets()

	return Controllers{
		api,
		pages,
		auth,
		dashboard,
		registration,
		assets,
		Fragments{},
	}
}

func redirectHx(w http.ResponseWriter, url string) error {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusSeeOther)

	return nil
}

func redirect(
	w http.ResponseWriter,
	r *http.Request,
	url string,
) error {
	http.Redirect(w, r, url, http.StatusSeeOther)
	return nil
}

// func adminOnlyAction(
// 	c echo.Context,
// 	dbtx *pgxpool.Pool,
// ) error {
// 	appCtx := reqmeta.ExtractApp(c.Request().Context())
//
// 	actor, err := models.GetUser(c.Request().Context(), dbtx, appCtx.UserID)
// 	if err != nil {
// 		return err
// 	}
//
// 	if !actor.IsAdmin {
// 		return errors.New("user must be admin to perform this action")
// 	}
//
// 	return nil
// }
