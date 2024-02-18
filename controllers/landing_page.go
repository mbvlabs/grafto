package controllers

import (
	"github.com/mbv-labs/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) LandingPage(ctx echo.Context) error {
	return views.HomePage().Render(views.ExtractRenderDeps(ctx))
}
