package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/views"
)

type App struct {
	db    psql.Postgres
	cache otter.CacheWithVariableTTL[string, templ.Component]
}

func newApp(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
) App {
	return App{db, cache}
}

func (a App) LandingPage(ctx echo.Context) error {
	return views.HomePage().Render(renderArgs(ctx))
}

func (a App) AboutPage(ctx echo.Context) error {
	return views.AboutPage().Render(renderArgs(ctx))
}

func (a App) Redirect(ctx echo.Context) error {
	to := ctx.QueryParam("to")

	return redirectHx(ctx.Response(), to)
}
