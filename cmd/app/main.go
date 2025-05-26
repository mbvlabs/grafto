package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/a-h/templ"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/handlers/middleware"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/psql/queue"
	"github.com/mbvlabs/grafto/psql/queue/workers"
	"github.com/mbvlabs/grafto/router"
	"github.com/mbvlabs/grafto/server"
	"github.com/mbvlabs/grafto/telemetry"
	"riverqueue.com/riverui"
)

var appVersion string

func run(ctx context.Context) error {
	cfg := config.NewConfig()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	tel, err := telemetry.New(
		ctx,
		appVersion,
		&telemetry.StdoutExporter{
			LogLevel:   slog.LevelDebug,
			WithTraces: true,
		},
		&telemetry.NoopTraceExporter{},
		&telemetry.NoopMetricExporter{},
	)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer func() {
		if err := tel.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown telemetry", "error", err)
		}
	}()

	if err := telemetry.SetupRuntimeMetricsInCallback(telemetry.GetMeter()); err != nil {
		return fmt.Errorf("failed to setup callback metrics: %w", err)
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

	queueLogger, queueLoggerShutdown := telemetry.NewLogger(
		ctx,
		&telemetry.StdoutExporter{
			LogLevel:   slog.LevelError,
			WithTraces: true,
		},
	)
	defer func() {
		if err := queueLoggerShutdown(ctx); err != nil {
			slog.Error("Failed to shutdown telemetry", "error", err)
		}
	}()

	psql := psql.NewPostgres(conn, nil)
	psql.NewQueue(
		queue.WithLogger(queueLogger),
		queue.WithWorkers(queueWorkers),
	)

	opts := &riverui.ServerOpts{
		Client: psql.Queue(),
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

	if err := riverUI.Start(ctx); err != nil {
		return err
	}

	emailClient := clients.NewEmail()

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
		emailClient,
	)

	mw, err := middleware.New(tel.AppTracerProvider)
	if err != nil {
		return err
	}

	routes := router.New(
		ctx,
		handlers,
		mw,
		riverUI,
		tel.AppTracerProvider,
	)

	router, c := routes.SetupRoutes(ctx)

	server := server.NewHttp(c, router)

	if err := psql.Queue().Start(ctx); err != nil {
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
