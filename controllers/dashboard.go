package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) DashboardIndex(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "dashboard/index", views.RenderOpts{
		Layout: views.DashboardLayout,
		Data:   nil,
	})
}
