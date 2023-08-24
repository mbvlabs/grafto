package main

import (
	"net/http"

	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	v := views.NewViews()

	e := echo.New()

	e.Renderer = v

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home/index", views.RenderOpts{
			Data: "this is a long string",
		})
	})

	e.GET("/dashboard", func(c echo.Context) error {
		return c.Render(http.StatusOK, "dashboard/index", views.RenderOpts{
			Layout: views.DashboardLayout,
			Data:   nil,
		})
	})

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
