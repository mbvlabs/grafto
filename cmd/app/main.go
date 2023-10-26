package main

import (
	"os"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/pkg/tokens"
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

	logger := telemetry.SetupLogger()

	slog.SetDefault(logger)

	// Middleware
	router.Use(slogecho.New(logger))
	router.Use(middleware.Recover())

	conn := database.SetupDatabaseConnection(config.GetDatabaseURL())
	db := database.New(conn)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	q := queue.NewQueue(db)

	tokenManager := tokens.NewManager()

	controllers := controllers.NewController(*db, mailClient, v, *tokenManager, *q)

	server := routes.NewServer(router, v, controllers, logger)

	server.Start()
}
