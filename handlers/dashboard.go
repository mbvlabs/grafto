package handlers

import (
	"github.com/labstack/echo/v4"
	dashboards "github.com/mbvlabs/grafto/views/dashboards"
)

type Dashboards struct{}

func newDashboards() Dashboards {
	return Dashboards{}
}

func (d Dashboards) Index(ctx echo.Context) error {
	return dashboards.Home().Render(renderArgs(ctx))
}
