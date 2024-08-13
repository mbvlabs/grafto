package main

import (
	"context"
	"log/slog"

	"github.com/gorilla/sessions"
	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/http"
	"github.com/mbv-labs/grafto/http/handlers"
	mw "github.com/mbv-labs/grafto/http/middleware"
	"github.com/mbv-labs/grafto/models"
	awsses "github.com/mbv-labs/grafto/pkg/aws_ses"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/psql"
	"github.com/mbv-labs/grafto/psql/database"
	"github.com/mbv-labs/grafto/queue"
	"github.com/mbv-labs/grafto/routes"
	"github.com/mbv-labs/grafto/services"
)

var appRelease string

func main() {
	cfg := config.NewConfig()

	otel := telemetry.NewOtel(cfg)
	defer func() {
		if err := otel.Shutdown(); err != nil {
			panic(err)
		}
	}()

	appTracer := otel.NewTracer("app/tracer")

	client := telemetry.NewTelemetry(cfg, appRelease)
	if client != nil {
		defer client.Stop()
	}

	conn, err := psql.CreatePooledConnection(
		context.Background(),
		cfg.GetDatabaseURL(),
	)
	if err != nil {
		panic(err)
	}

	db := database.New(conn)
	psql := psql.NewPostgres(conn)
	riverClient := queue.NewClient(conn, queue.WithLogger(slog.Default()))

	authSessionStore := sessions.NewCookieStore(
		[]byte(cfg.SessionKey),
		[]byte(cfg.SessionEncryptionKey),
	)

	awsSes := awsses.New()

	authSvc := services.NewAuth(psql, authSessionStore, cfg)
	tokenService := services.NewTokenSvc(psql, cfg.TokenSigningKey)
	emailService := services.NewEmailSvc(cfg, &awsSes, riverClient)

	userModelSvc := models.NewUserService(psql, authSvc)

	flashStore := handlers.NewCookieStore("")
	baseHandler := handlers.NewDependencies(cfg, db, flashStore, riverClient, appTracer)
	appHandlers := handlers.NewApp(baseHandler)
	dashboardHandlers := handlers.NewDashboard(baseHandler)
	registrationHandlers := handlers.NewRegistration(
		authSvc,
		baseHandler,
		userModelSvc,
		*tokenService,
		emailService,
	)
	apiHandlers := handlers.NewApi()
	authenticationHandlers := handlers.NewAuthentication(
		authSvc,
		baseHandler,
		userModelSvc,
		*tokenService,
		emailService,
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

	router := routes.SetupRoutes()

	server := http.NewServer(router, cfg)

	server.Start()
}
