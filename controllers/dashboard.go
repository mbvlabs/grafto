package controllers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/views"
)

func (c *Controller) DashboardIndex(ctx echo.Context) error {
	return views.DashboardPage().Render(views.ExtractRenderDeps(ctx))
}
