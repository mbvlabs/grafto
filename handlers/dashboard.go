package handlers

import (
	"github.com/labstack/echo/v4"
	dashboardViews "github.com/mbvlabs/grafto/views/dashboards"
)

type Dashboard struct{}

func newDashboard() Dashboard {
	return Dashboard{}
}

func (d Dashboard) Index(ctx echo.Context) error {
	return dashboardViews.Home().Render(renderArgs(ctx))
}
