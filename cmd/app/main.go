package main

import (
	"os"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/routes"
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

func main() {
	v := views.NewViews()

	router := echo.New()
	router.Renderer = v

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Middleware
	router.Use(slogecho.New(logger))

	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	conn := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	db := database.New(conn)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))

	mailClient := mail.NewMail(&postmark)
	controllers := controllers.NewController(*db, mailClient)

	server := routes.NewServer(router, v, controllers)

	server.Start()
}
