package views

import (
	"context"
	"io"

	"github.com/labstack/echo/v4"
)

func ExtractRenderDeps(ctx echo.Context) (context.Context, io.Writer) {
	return ctx.Request().Context(), ctx.Response().Writer
}
