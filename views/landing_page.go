package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/pages"
	"github.com/labstack/echo/v4"
)

func LandingPage(ctx echo.Context) error {
	return layouts.Base(pages.LandingPage()).Render(extractRenderDeps(ctx))
}
