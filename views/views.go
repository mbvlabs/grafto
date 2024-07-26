package views

import (
	"context"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/http/middleware"
)

func setUserCtx(ctx echo.Context) context.Context {
	userCtx := ctx.(*middleware.UserContext)
	return context.WithValue(ctx.Request().Context(), middleware.UserContext{}, userCtx)
}

// ExtractRenderDeps extracts the context and writer from the echo context and sets the user context
func ExtractRenderDeps(ctx echo.Context) (context.Context, io.Writer) {
	return setUserCtx(ctx), ctx.Response().Writer
}
