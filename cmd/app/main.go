package main

import (
	"os"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/routes"
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	v := views.NewViews()

	router := echo.New()
	router.Renderer = v

	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	conn := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	db := database.New(conn)
	controllers := controllers.NewController(*db)

	server := routes.NewServer(router, v, controllers)

	server.Start()
}
