package handlers

import (
	"errors"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/views"
)

type App struct {
	Base
}

func NewApp(base Base) App {
	return App{base}
}

func (a *App) LandingPage(ctx echo.Context) error {
	slog.ErrorContext(ctx.Request().Context(), "yo mate", "error", errors.New("jdklsajdlksajldksa"))
	return views.HomePage().Render(views.ExtractRenderDeps(ctx))
}
