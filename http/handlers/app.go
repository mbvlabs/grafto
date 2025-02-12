package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/views"
)

const landingPageCacheKey = "LandingPage"

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

func (a *App) LandingPage(ctx echo.Context) error {
	// if value, ok := a.cache.Get(landingPageCacheKey); ok {
	// 	slog.Info(
	// 		"$$$$$$$$$$$$$$$$$$$$$$$$$ CACHE HIT $$$$$$$$$$$$$$$$$$$$$$$$$$",
	// 	)
	// 	return views.HomePage(value).Render(renderArgs(ctx))
	// }

	// var sb strings.Builder
	// if err := views.Home()
	// 	log.Fatalf("failed to render to string: %v", err)
	// }
	//
	// cachedComponent := views.Home()
	//
	// if ok := a.cache.Set(landingPageCacheKey, cachedComponent, time.Hour*time.Duration(24)); !ok {
	// 	return views.HomePage(cachedComponent).Render(renderArgs(ctx))
	// }

	return views.HomePage().Render(renderArgs(ctx))
}

func (a *App) AboutPage(ctx echo.Context) error {
	return views.AboutPage().Render(renderArgs(ctx))
}
