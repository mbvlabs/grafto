package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
)

func (c *Controller) HomeIndex(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "home/index", views.RenderOpts{
		Data: nil,
	})
}
