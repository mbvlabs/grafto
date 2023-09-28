package middleware

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/labstack/echo/v4"
)

func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authenticated, err := services.IsAuthenticated(c.Request())
		if err == nil {
			telemetry.Logger.Error("could not get authenticated status", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		if authenticated {
			return next(c)
		} else {
			return c.Redirect(http.StatusPermanentRedirect, "/")
		}
	}
}
