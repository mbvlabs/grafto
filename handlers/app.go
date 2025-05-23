package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/routes"
	"github.com/mbvlabs/grafto/telemetry"
	"github.com/mbvlabs/grafto/views"
)

type App struct {
	db     psql.Postgres
	cache  otter.CacheWithVariableTTL[string, templ.Component]
	logger *telemetry.Logger
}

func newApp(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
	logger *telemetry.Logger,
) App {
	return App{db, cache, logger}
}

func (a App) LandingPage(c echo.Context) error {
	a.logger.InfoContext(c.Request().Context(), "yoyoyoy")
	return views.HomePage().Render(renderArgs(c))
}

func (a App) AboutPage(c echo.Context) error {
	return views.AboutPage().Render(renderArgs(c))
}

func (a App) Redirect(c echo.Context) error {
	to := c.QueryParam("to")
	for _, r := range routes.AllRoutes {
		if to == r.Path {
			return redirectHx(c.Response(), to)
		}
	}

	a.logger.WithContext(c.Request().Context()).Info(
		"security warning: someone tried to missue open redirect",
		"to", to,
		"ip", c.RealIP(),
	)
	return redirect(c.Response(), c.Request(), "/")
}
