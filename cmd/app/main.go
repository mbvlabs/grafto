package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/a-h/templ"
	"github.com/lmittmann/tint"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/queue"
	"github.com/mbvlabs/grafto/queue/workers"
	"github.com/mbvlabs/grafto/routes"
	"github.com/mbvlabs/grafto/server"
	"github.com/mbvlabs/grafto/services"
	"riverqueue.com/riverui"
)

func developmentLogger() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	)
}

func queueLogger() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelError,
			TimeFormat: time.Kitchen,
		}),
	)
}

func productionLogger() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelError,
			TimeFormat: time.Kitchen,
		}),
	)
}

func run(ctx context.Context) error {
	cfg := config.NewConfig()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if cfg.Environment == config.DEV_ENVIRONMENT {
		slog.SetDefault(developmentLogger())
	}
	if cfg.Environment == config.PROD_ENVIRONMENT {
		slog.SetDefault(productionLogger())
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
	riverClient := queue.NewClient(
		conn,
		queue.WithLogger(queueLogger()),
		queue.WithWorkers(queueWorkers),
	)
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

	emailSvc := services.NewEmail()

	cacheBuilder, err := otter.NewBuilder[string, templ.Component](20)
	if err != nil {
		return err
	}

	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
	if err != nil {
		return err
	}

	handlers := handlers.NewHandlers(
		psql,
		pageCacher,
		emailSvc,
	)

	routes := routes.NewRoutes(
		handlers,
		riverUI,
	)

	router, c := routes.SetupRoutes(ctx)

	server := server.NewHttp(c, router)

	if err := riverClient.Start(ctx); err != nil {
		return err
	}
	return server.Start(c)
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
