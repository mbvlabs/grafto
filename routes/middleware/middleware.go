package middleware

import (
	"net/http"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ContextUserID struct {
	echo.Context
	userID uuid.UUID
}

func (c *ContextUserID) GetID() uuid.UUID {
	return c.userID
}

func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authenticated, userID, err := services.IsAuthenticated(c.Request())
		if err != nil {
			telemetry.Logger.Error("could not get authenticated status", "error", err)
			return c.Redirect(http.StatusPermanentRedirect, "/500")
		}

		if authenticated {
			ctx := &ContextUserID{c, userID}
			return next(ctx)
		} else {
			return c.Redirect(http.StatusPermanentRedirect, "/")
		}
	}
}
