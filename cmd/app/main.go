package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/controllers"
	emails "github.com/mbvlabs/grafto/pkg/email_client"
	"github.com/mbvlabs/grafto/pkg/telemetry"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/queue"
	"github.com/mbvlabs/grafto/queue/workers"
	"github.com/mbvlabs/grafto/routes"
	"github.com/mbvlabs/grafto/server"
	mw "github.com/mbvlabs/grafto/server/middleware"
	"github.com/mbvlabs/grafto/services"
	"riverqueue.com/riverui"
)

var appRelease string

func run(ctx context.Context) error {
	cfg := config.NewConfig()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	otel := telemetry.NewOtel()
	defer func() {
		if err := otel.Shutdown(); err != nil {
			panic(err)
		}
	}()

	// appTracer := otel.NewTracer("app/tracer")

	client := telemetry.NewTelemetry(cfg, appRelease, strings.ToLower(cfg.ProjectName))
	if client != nil {
		defer client.Stop()
	}

	conn, err := psql.CreatePooledConnection(
		ctx,
		cfg.GetDatabaseURL(),
	)
	if err != nil {
		return err
	}
	queueWorkers, err := workers.SetupWorkers(workers.WorkerDependencies{})
	if err != nil {
		return err
	}
	riverClient := queue.NewClient(conn, queue.WithLogger(slog.Default()), queue.WithWorkers(queueWorkers))
	psql := psql.NewPostgres(conn, riverClient)

	opts := &riverui.ServerOpts{
		Client: riverClient,
		DB:     conn,
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.SetLogLoggerLevel(slog.LevelError),
		})),
		Prefix: "/river",
	}
	riverUI, err := riverui.NewServer(opts)
	if err != nil {
		return err
	}
	// Start the server to initialize background processes for caching and periodic queries:
	if err := riverUI.Start(ctx); err != nil {
		return err
	}

	awsSES := emails.NewSESClient()
	emailClient := emails.NewEmail(awsSES)

	authSvc := services.NewAuthNew(psql, emailClient)
	tokenService := services.NewTokenSvc(psql, cfg.TokenSigningKey)

	cacheBuilder, err := otter.NewBuilder[string, string](20)
	if err != nil {
		return err
	}

	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
	if err != nil {
		return err
	}

	appHandlers := controllers.NewApp(psql, pageCacher)
	dashboardHandlers := controllers.NewDashboard()
	registrationHandlers := controllers.NewRegistration(
		authSvc,
		psql,
		*tokenService,
		emailClient,
	)
	apiHandlers := controllers.NewApi()
	authenticationHandlers := controllers.NewAuthentication(
		authSvc,
		psql,
		*tokenService,
		emailClient,
	)

	middleware := mw.NewMiddleware()

	routes := routes.NewRoutes(
		appHandlers,
		dashboardHandlers,
		authenticationHandlers,
		registrationHandlers,
		apiHandlers,
		middleware,
		riverUI,
	)

	router := routes.SetupRoutes()

	server := server.NewServer(ctx, router)

	return server.Start(ctx)
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
