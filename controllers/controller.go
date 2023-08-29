package controllers

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db   database.Queries
	mail mail.Mail
}

func NewController(db database.Queries, mail mail.Mail) Controller {
	return Controller{
		db,
		mail,
	}
}

func (c *Controller) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []byte("app is healthy and running"))
}
