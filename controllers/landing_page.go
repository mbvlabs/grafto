package controllers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/views"
)

func (c *Controller) LandingPage(ctx echo.Context) error {
	return views.HomePage().Render(views.ExtractRenderDeps(ctx))
}
