package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/pages"
	"github.com/labstack/echo/v4"
)

func HomeIndex(ctx echo.Context) error {
	return layouts.Base(pages.HomeIndex()).Render(extractRenderDeps(ctx))
}
