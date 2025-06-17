package handlers

import (
	"github.com/labstack/echo/v4"
	dashboards "github.com/mbvlabs/grafto/views/dashboards"
)

type Dashboard struct{}

func newDashboard() Dashboard {
	return Dashboard{}
}

func (d Dashboard) Index(ctx echo.Context) error {
	return dashboards.Home().Render(renderArgs(ctx))
}
