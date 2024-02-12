package main

import (
	"context"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/pkg/tokens"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/routes"
	"github.com/MBvisti/grafto/server"
	mw "github.com/MBvisti/grafto/server/middleware"
	"github.com/MBvisti/grafto/services"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

func main() {
	router := echo.New()

	logger := telemetry.SetupLogger()
	slog.SetDefault(logger)

	cfg := config.New()

	// Middleware
	router.Use(slogecho.New(logger))
	router.Use(middleware.Recover())

	conn := database.SetupDatabasePool(context.Background(), cfg.Db.GetUrlString())
	db := database.New(conn)

	// q := queue.New(db)
	// if err := q.InitilizeRepeatingJobs(context.Background(), nil); err != nil {
	// 	panic(err)
	// }

	postmark := mail.NewPostmark(cfg.ExternalProviders.PostmarkApiToken)

	mailClient := mail.NewMail(&postmark)
	tokenManager := tokens.NewManager(cfg.Auth.TokenSigningKey)

	authSessionStore := sessions.NewCookieStore([]byte(cfg.Auth.SessionKey), []byte(cfg.Auth.SessionEncryptionKey))

	services := services.NewServices(authSessionStore)

	controllers := controllers.NewController(*db, mailClient, *tokenManager, cfg, services)

	serverMW := mw.NewMiddleware(services)

	routes := routes.NewRoutes(controllers, serverMW, cfg)
	router = routes.SetupRoutes()

	server := server.NewServer(router, controllers, logger, cfg, services)

	server.Start()
}
