package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/views"
)

type App struct {
	Base
}

func NewApp(base Base) App {
	return App{base}
}

func (a *App) LandingPage(ctx echo.Context) error {
	return views.HomePage().Render(views.ExtractRenderDeps(ctx))
}
