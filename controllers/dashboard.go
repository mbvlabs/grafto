package controllers

import (
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) DashboardIndex(ctx echo.Context) error {
	return views.DashboardPage().Render(views.ExtractRenderDeps(ctx))
}
