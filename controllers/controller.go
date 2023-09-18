package controllers

import (
	"net/http"
	"os"
	"strings"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/views"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db       database.Queries
	mail     mail.Mail
	views    views.Views
	validate *validator.Validate
}

func NewController(db database.Queries, mail mail.Mail, views views.Views) Controller {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return Controller{
		db,
		mail,
		views,
		validate,
	}
}

func (c *Controller) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []byte("app is healthy and running"))
}

func (c *Controller) InternalError(ctx echo.Context) error {
	domainName := os.Getenv("DOMAIN_NAME")
	referere := strings.Split(ctx.Request().Referer(), domainName)

	var from string
	if len(referere) == 1 || referere[1] == "" {
		from = "/"
	} else {
		from = referere[1]
	}

	return c.views.InternalServerErr(ctx, views.InternalServerErrData{
		FromLocation: from,
	})
}
