package main

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mbv-labs/grafto/controllers"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/config"
	"github.com/mbv-labs/grafto/pkg/mail"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/repository/psql"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/mbv-labs/grafto/routes"
	"github.com/mbv-labs/grafto/server"
	mw "github.com/mbv-labs/grafto/server/middleware"
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

	postmark := mail.NewPostmark(cfg.ExternalProviders.PostmarkApiToken)
	mailClient := mail.NewMail(&postmark)

	tokenManager := tokens.NewManager(cfg.Auth.TokenSigningKey)

	authSessionStore := sessions.NewCookieStore(
		[]byte(cfg.Auth.SessionKey),
		[]byte(cfg.Auth.SessionEncryptionKey),
	)

	validator := validator.New()
	validator.RegisterStructValidation(models.PasswordMatchValidation, models.NewUserValidation{})

	authSvc := services.NewAuth(psql, authSessionStore, cfg)
	userModelSvc := models.NewUserService(psql, authSvc, validator)

	riverClient := queue.NewClient(conn, queue.WithLogger(logger))

	controllers := controllers.NewController(
		*db,
		mailClient,
		userModelSvc,
		authSvc,
		*tokenManager,
		cfg,
		riverClient,
		authSessionStore,
	)

	serverMW := mw.NewMiddleware(authSvc)

	routes := routes.NewRoutes(controllers, serverMW, cfg)
	router = routes.SetupRoutes()

	server := server.NewServer(router, controllers, logger, cfg)

	server.Start()
}
