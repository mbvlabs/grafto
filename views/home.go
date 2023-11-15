package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/templates"
	"github.com/labstack/echo/v4"
)

func HomeIndex(ctx echo.Context) error {
	return layouts.Base(templates.HomeIndex()).Render(extractRenderDeps(ctx))
}
