package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/views"
)

type Dashboard struct {
	Base
}

func NewDashboard(base Base) Dashboard {
	return Dashboard{base}
}

func (d *Dashboard) Index(ctx echo.Context) error {
	return views.DashboardPage().Render(views.ExtractRenderDeps(ctx))
}
