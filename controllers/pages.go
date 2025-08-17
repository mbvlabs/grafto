package controllers

import (
	"log/slog"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/routes"
	"github.com/mbvlabs/grafto/views"
)

type Pages struct {
	db    psql.Postgres
	cache otter.CacheWithVariableTTL[string, templ.Component]
}

func newPages(
	db psql.Postgres,
	cache otter.CacheWithVariableTTL[string, templ.Component],
) Pages {
	return Pages{db, cache}
}

func (p Pages) LandingPage(c echo.Context) error {
	return views.HomePage().Render(renderArgs(c))
}

func (p Pages) AboutPage(c echo.Context) error {
	return views.AboutPage().Render(renderArgs(c))
}

func (p Pages) Redirect(c echo.Context) error {
	to := c.QueryParam("to")
	for _, r := range routes.AllRoutes {
		if to == r.Path {
			return redirectHx(c.Response(), to)
		}
	}

	slog.InfoContext(c.Request().Context(),
		"security warning: someone tried to missue open redirect",
		"to", to,
		"ip", c.RealIP(),
	)

	return redirect(c.Response(), c.Request(), "/")
}

func (p Pages) NotFoundPage(c echo.Context) error {
	c.Response().Status = 404
	return views.NotFoundPage().Render(renderArgs(c))
}
