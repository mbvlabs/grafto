package main

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mbvlabs/grafto/views"
)

func main() {
	e := echo.New()

	// Serve static assets
	e.Static("/assets", "assets")

	// Add CORS middleware for better development experience
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return views.ComponentsShowcase().Render(c.Request().Context(), c.Response().Writer)
	})

	slog.Info("starting the components server on port: 5555")
	e.Logger.Fatal(e.Start(":5555"))
}
