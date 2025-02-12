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
	// if value, ok := a.cache.Get(landingPageCacheKey); ok {
	// 	return views.HomePage(value).Render(renderArgs(ctx))
	// }
	//
	// var sb strings.Builder
	// if err := views.Home().Render(ctx.Request().Context(), &sb); err != nil {
	// 	log.Fatalf("failed to render to string: %v", err)
	// }
	//
	// cachedHtml := sb.String()
	//
	// if ok := a.cache.Set(landingPageCacheKey, cachedHtml, time.Hour*time.Duration(24)); !ok {
	// 	return views.HomePage(cachedHtml).Render(renderArgs(ctx))
	// }
	return views.HomePage().Render(renderArgs(ctx))
}
