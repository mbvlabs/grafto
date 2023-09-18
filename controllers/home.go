package controllers

import (
	"github.com/labstack/echo/v4"
)

func (c *Controller) HomeIndex(ctx echo.Context) error {
	return c.views.Home(ctx)
}
