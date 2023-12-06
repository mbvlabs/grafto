package views

import (
	"github.com/MBvisti/grafto/views/internal/layouts"
	"github.com/MBvisti/grafto/views/internal/pages"
	"github.com/labstack/echo/v4"
)

func Dashboard(ctx echo.Context) error {
	return layouts.Dashboard(pages.DashboardIndex()).Render(extractRenderDeps(ctx))
}
