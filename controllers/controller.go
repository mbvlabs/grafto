package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/config"
	"github.com/mbv-labs/grafto/pkg/mail"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/mbv-labs/grafto/services"
	"github.com/mbv-labs/grafto/views"
	"github.com/riverqueue/river"
)

type Controller struct {
	db               database.Queries
	mail             mail.Mail
	userModel        models.UserService
	authSvc          services.Auth
	validate         *validator.Validate
	tknManager       tokens.Manager
	cfg              config.Cfg
	queueClient      *river.Client[pgx.Tx]
	authSessionStore *sessions.CookieStore
}

func NewController(
	db database.Queries,
	mail mail.Mail,
	userModel models.UserService,
	authSvc services.Auth,
	tknManager tokens.Manager,
	cfg config.Cfg,
	qc *river.Client[pgx.Tx],
	authSessionStore *sessions.CookieStore,
) Controller {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return Controller{
		db,
		mail,
		userModel,
		authSvc,
		validate,
		tknManager,
		cfg,
		qc,
		authSessionStore,
	}
}

func (c *Controller) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "app is healthy and running")
}

func (c *Controller) InternalError(ctx echo.Context) error {
	referere := strings.Split(ctx.Request().Referer(), c.cfg.App.ServerHost)

	var from string
	if len(referere) == 1 || referere[1] == "" {
		from = "/"
	} else {
		from = referere[1]
	}

	return views.InternalServerError(views.Head{
		Title:       "Internal Server Error",
		Description: "An error occurred while processing your request",
	}, from).Render(views.ExtractRenderDeps(ctx))
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
