package main

import (
	"net/http"
	"os"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/repository/database"
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
	conn := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	db := database.New(conn)
	controllers := controllers.NewController(*db)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home/index", nil)
	})

	e.GET("/dashboard", func(c echo.Context) error {
		return c.Render(http.StatusOK, "dashboard/index", views.RenderOpts{
			Layout: views.DashboardLayout,
			Data:   nil,
		})
	})

	e.GET("/health", func(c echo.Context) error {
		return controllers.Health(c)
	})

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
