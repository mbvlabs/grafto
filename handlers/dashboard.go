package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/views/dashboard"
)

type Dashboard struct{}

func newDashboard() Dashboard {
	return Dashboard{}
}

func (d Dashboard) Index(ctx echo.Context) error {
	return dashboard.Home().Render(renderArgs(ctx))
}
