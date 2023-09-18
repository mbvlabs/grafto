package views

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (v Views) Home(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "home/index", RenderOpts{
		Layout: BaseLayout,
	})
}
