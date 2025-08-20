package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct{}

func newAPI() API {
	return API{}
}

func (a API) AppHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "app is healthy and running")
}
