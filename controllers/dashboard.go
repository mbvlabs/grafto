package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) DashboardIndex(ctx echo.Context) error {

	contextUserID := ctx.(*middleware.ContextUserID)
	telemetry.Logger.Info("ctx", "id", contextUserID.GetID())

	return ctx.Render(http.StatusOK, "dashboard/index", views.RenderOpts{
		Layout: views.DashboardLayout,
		Data:   nil,
	})
}
