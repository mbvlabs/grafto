package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/views"
)

const landingPageCacheKey = "LandingPage"

type App struct {
	db    psql.Postgres
	cache otter.CacheWithVariableTTL[string, string]
}

func newApp(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, string],
) App {
	return App{db, cache}
}

func (a *App) LandingPage(ctx echo.Context) error {
	return views.HomePage().Render(extractRenderDeps(ctx))
}

func (a *App) AboutPage(ctx echo.Context) error {
	return views.AboutPage().Render(extractRenderDeps(ctx))
}
