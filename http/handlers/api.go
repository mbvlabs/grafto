package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Api struct{}

func NewApi() Api {
	return Api{}
}

func (a *Api) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "app is healthy and running")
}
