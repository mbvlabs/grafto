package views

import (
	"context"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/server/middleware"
)

func setAppCtx(ctx echo.Context) context.Context {
	appCtx := ctx.Get(middleware.AppContextName)
	return context.WithValue(
		ctx.Request().Context(),
		middleware.AppContext{},
		appCtx,
	)
}

// ExtractRenderDeps extracts the context and writer from the echo context and sets the user context
func ExtractRenderDeps(ctx echo.Context) (context.Context, io.Writer) {
	return setAppCtx(ctx), ctx.Response().Writer
}
