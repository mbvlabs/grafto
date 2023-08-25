package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/repository/database"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db database.Queries
}

func NewController(db database.Queries) Controller {
	return Controller{
		db,
	}
}

func (c *Controller) Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []byte("app is healthy and running"))
}
