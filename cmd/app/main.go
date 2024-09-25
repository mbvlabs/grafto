package main

import (
	"context"
	"log/slog"

	"github.com/gorilla/sessions"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/http"
	"github.com/mbvlabs/grafto/http/handlers"
	mw "github.com/mbvlabs/grafto/http/middleware"
	"github.com/mbvlabs/grafto/models"
	awsses "github.com/mbvlabs/grafto/pkg/aws_ses"
	"github.com/mbvlabs/grafto/pkg/telemetry"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/queue"
	"github.com/mbvlabs/grafto/routes"
	"github.com/mbvlabs/grafto/services"
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

	client := telemetry.NewTelemetry(cfg, appRelease, "grafto")
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
	baseHandler := handlers.NewDependencies(
		cfg,
		psql,
		flashStore,
		riverClient,
		appTracer,
	)
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
