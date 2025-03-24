package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
)

func registrationRoutes(
	router *echo.Echo,
	handlers handlers.Registration,
) {
	router.GET(paths.CreateUser.URL, func(c echo.Context) error {
		return handlers.CreateUser(c)
	}).Name = paths.CreateUser.Name
	router.POST(paths.StoreUser.URL, func(c echo.Context) error {
		return handlers.StoreUser(c)
	}).Name = paths.StoreUser.Name

	router.GET(paths.VerifyEmail.URL, func(c echo.Context) error {
		return handlers.VerifyUserEmail(c)
	}).Name = paths.VerifyEmail.Name
}
