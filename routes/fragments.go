package routes

import (
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/routes/paths"
	"github.com/mbvlabs/grafto/views/fragments"
)

func fragmentRoutes(router *echo.Echo) {
	router.GET("/load-csrf", func(c echo.Context) error {
		return fragments.CsrfToken(csrf.Token(c.Request())).
			Render(c.Request().Context(), c.Response())
	}).Name = paths.LoadCsrfToken.String()
}
