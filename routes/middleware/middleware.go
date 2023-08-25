package middleware

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthOnly is only a placeholder for now
func AuthOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		rando := rand.Float64()

		if rando > 0.5 {
			return next(c)
		}

		req := c.Request()
		log.Printf("ref: %v", req.Referer())
		return c.Redirect(http.StatusPermanentRedirect, "/dashboard")
	}
}
