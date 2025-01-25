package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/views"
)

type Dashboard struct{}

func newDashboard() Dashboard {
	return Dashboard{}
}

func (d *Dashboard) Index(ctx echo.Context) error {
	return views.DashboardPage().Render(extractRenderDeps(ctx))
}
