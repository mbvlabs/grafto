package main

import (
	"context"
	"log/slog"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mbv-labs/grafto/controllers"
	"github.com/mbv-labs/grafto/pkg/config"
	"github.com/mbv-labs/grafto/pkg/mail"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/repository/database"
	"github.com/mbv-labs/grafto/routes"
	"github.com/mbv-labs/grafto/server"
	mw "github.com/mbv-labs/grafto/server/middleware"
	slogecho "github.com/samber/slog-echo"
)

func main() {
	ctx := context.Background()
	router := echo.New()

	logger := telemetry.SetupLogger()
	slog.SetDefault(logger)

	cfg := config.New()

	// Middleware
	router.Use(slogecho.New(logger))
	router.Use(middleware.Recover())

	conn := database.SetupDatabasePool(context.Background(), cfg.Db.GetUrlString())
	db := database.New(conn)

	postmark := mail.NewPostmark(cfg.ExternalProviders.PostmarkApiToken)
	mailClient := mail.NewMail(&postmark)

	tokenManager := tokens.NewManager(cfg.Auth.TokenSigningKey)

	authSessionStore := sessions.NewCookieStore(
		[]byte(cfg.Auth.SessionKey),
		[]byte(cfg.Auth.SessionEncryptionKey),
	)

	queueDbPool, err := pgxpool.New(context.Background(), cfg.Db.GetQueueUrlString())
	if err != nil {
		panic(err)
	}

	if err := queueDbPool.Ping(ctx); err != nil {
		panic(err)
	}

	riverClient := queue.NewClient(queueDbPool, queue.WithLogger(logger))

	controllers := controllers.NewController(
		*db,
		mailClient,
		*tokenManager,
		cfg,
		riverClient,
		authSessionStore,
	)

	serverMW := mw.NewMiddleware(authSessionStore)

	routes := routes.NewRoutes(controllers, serverMW, cfg)
	router = routes.SetupRoutes()

	server := server.NewServer(router, controllers, logger, cfg)

	server.Start()
}
