package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/a-h/templ"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/psql/queue"
	"github.com/mbvlabs/grafto/psql/queue/workers"
	"github.com/mbvlabs/grafto/router"
	"github.com/mbvlabs/grafto/router/middleware"
	"github.com/mbvlabs/grafto/telemetry"
	"golang.org/x/sync/errgroup"
	"riverqueue.com/riverui"
)

var appVersion string

func startServer(ctx context.Context, srv *http.Server, env string) error {
	if env == config.PROD_ENVIRONMENT {
		eg, egCtx := errgroup.WithContext(ctx)

		eg.Go(func() error {
			if err := srv.ListenAndServe(); err != nil &&
				err != http.ErrServerClosed {
				return fmt.Errorf("server error: %w", err)
			}
			return nil
		})

		eg.Go(func() error {
			<-egCtx.Done()
			slog.Info("initiating graceful shutdown")
			shutdownCtx, cancel := context.WithTimeout(
				ctx,
				10*time.Second,
			)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown error: %w", err)
			}
			return nil
		})

		if err := eg.Wait(); err != nil {
			slog.Info("wait error", "e", err)
			return err
		}

		return nil
	}

	return srv.ListenAndServe()
}

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
		cfg.DB.GetDatabaseURL(),
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

	routes, err := router.New(
		handlers,
		mw,
		riverUI,
		tel.AppTracerProvider,
	)
	if err != nil {
		return err
	}

	router := routes.SetupRoutes()

	if err := psql.Queue().Start(ctx); err != nil {
		return err
	}

	port := config.Cfg.App.ServerPort
	host := config.Cfg.App.ServerHost

	srv := &http.Server{
		Addr:         fmt.Sprintf("%v:%v", host, port),
		Handler:      router,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 5 * time.Second,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

	slog.InfoContext(ctx, "starting server", "host", host, "port", port)
	return startServer(ctx, srv)
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
