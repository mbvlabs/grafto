package main

import (
	"context"
	"log/slog"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mbv-labs/grafto/http"
	"github.com/mbv-labs/grafto/http/handlers"
	mw "github.com/mbv-labs/grafto/http/middleware"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/config"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/repository/psql"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/mbv-labs/grafto/routes"
	"github.com/mbv-labs/grafto/services"
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

	conn, err := psql.CreatePooledConnection(
		context.Background(),
		cfg.Db.GetUrlString(),
	)
	if err != nil {
		panic(err)
	}

	db := database.New(conn)
	psql := psql.NewPostgres(conn)

	// postmark := mail.NewPostmark(cfg.ExternalProviders.PostmarkApiToken)
	// mailClient := mail.NewMail(&postmark)

	tokenManager := tokens.NewManager(cfg.Auth.TokenSigningKey)

	authSessionStore := sessions.NewCookieStore(
		[]byte(cfg.Auth.SessionKey),
		[]byte(cfg.Auth.SessionEncryptionKey),
	)

	authSvc := services.NewAuth(psql, authSessionStore, cfg)
	userModelSvc := models.NewUserService(psql, authSvc)

	riverClient := queue.NewClient(conn, queue.WithLogger(logger))

	flashStore := handlers.NewCookieStore("")
	baseHandler := handlers.NewDependencies(cfg, db, flashStore, riverClient)
	appHandlers := handlers.NewApp(baseHandler)
	dashboardHandlers := handlers.NewDashboard(baseHandler)
	registrationHandlers := handlers.NewRegistration(
		authSvc,
		baseHandler,
		userModelSvc,
		*tokenManager,
	)
	apiHandlers := handlers.NewApi()
	authenticationHandlers := handlers.NewAuthentication(
		authSvc,
		baseHandler,
		userModelSvc,
		*tokenManager,
	)

	serverMW := mw.NewMiddleware(authSvc)

	routes := routes.NewRoutes(
		appHandlers,
		dashboardHandlers,
		authenticationHandlers,
		registrationHandlers,
		apiHandlers,
		baseHandler,
		serverMW,
		cfg,
	)
	router = routes.SetupRoutes()

	server := http.NewServer(router, logger, cfg)

	server.Start()
}
