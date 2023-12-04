package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/tokens"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/views"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	db         database.Queries
	mail       mail.Mail
	validate   *validator.Validate
	tknManager tokens.Manager
	queue      queue.Queue
}

func NewController(
	db database.Queries, mail mail.Mail, tknManager tokens.Manager, queue queue.Queue) Controller {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return Controller{
		db,
		mail,
		validate,
		tknManager,
		queue,
	}
}

func (c *Controller) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []byte("app is healthy and running"))
}

func (c *Controller) InternalError(ctx echo.Context) error {
	hostName := os.Getenv("HOST")
	referere := strings.Split(ctx.Request().Referer(), hostName)

	var from string
	if len(referere) == 1 || referere[1] == "" {
		from = "/"
	} else {
		from = referere[1]
	}

	return views.InternalServerErr(ctx, views.InternalServerErrData{
		FromLocation: from,
	})
}

func (c *Controller) Redirect(ctx echo.Context) error {
	toLocation := ctx.QueryParam("to")
	if toLocation == "" {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		return c.InternalError(ctx)
	}

	ctx.Response().Writer.Header().Add("HX-Redirect", fmt.Sprintf("/%s", toLocation))

	return nil
}
