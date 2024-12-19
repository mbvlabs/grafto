package controllers

import (
	"log"
	"net/http"
	"strings"
	"time"

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

func NewApp(db psql.Postgres, cache otter.CacheWithVariableTTL[string, string]) App {
	return App{db, cache}
}

func (a *App) LandingPage(ctx echo.Context) error {
	if value, ok := a.cache.Get(landingPageCacheKey); ok {
		return ctx.HTML(http.StatusOK, value)
	}

	var sb strings.Builder
	if err := views.HomePage().Render(ctx.Request().Context(), &sb); err != nil {
		log.Fatalf("failed to render to string: %v", err)
	}

	cachedHtml := sb.String()

	if ok := a.cache.Set(landingPageCacheKey, cachedHtml, time.Hour*time.Duration(24)); !ok {
		return views.HomePage().Render(ctx.Request().Context(), ctx.Response())
	}

	return ctx.HTML(http.StatusOK, cachedHtml)
}
