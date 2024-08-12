package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mbv-labs/grafto/config"
	awsses "github.com/mbv-labs/grafto/pkg/aws_ses"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/psql"
	"github.com/mbv-labs/grafto/psql/database"
	"github.com/mbv-labs/grafto/psql/queue"
	"github.com/mbv-labs/grafto/psql/queue/workers"
	"github.com/riverqueue/river"
)

var appRelease string

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()

	otel := telemetry.NewOtel(cfg)
	defer func() {
		if err := otel.Shutdown(); err != nil {
			panic(err)
		}
	}()

	workerTracer := otel.NewTracer("worker/tracer")

	client := telemetry.NewTelemetry(cfg, appRelease)
	if client != nil {
		defer client.Stop()
	}

	awsSes := awsses.New()

	conn, err := psql.CreatePooledConnection(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		panic(err)
	}
	db := database.New(conn)

	jobStarted := make(chan struct{})

	workers, err := workers.SetupWorkers(workers.WorkerDependencies{
		DB:      db,
		Emailer: awsSes,
		Tracer:  workerTracer,
	})
	if err != nil {
		panic(err)
	}

	q := map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 100}}
	riverClient := queue.NewClient(
		conn,
		queue.WithQueues(q),
		queue.WithWorkers(workers),
		queue.WithLogger(slog.Default()),
	)

	if err := riverClient.Start(ctx); err != nil {
		panic(err)
	}

	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)

	// This is meant to be a realistic-looking stop goroutine that might go in a
	// real program. It waits for SIGINT/SIGTERM and when received, tries to stop
	// gracefully by allowing a chance for jobs to finish. But if that isn't
	// working, a second SIGINT/SIGTERM will tell it to terminate with prejudice and
	// it'll issue a hard stop that cancels the context of all active jobs. In
	// case that doesn't work, a third SIGINT/SIGTERM ignores River's stop procedure
	// completely and exits uncleanly.
	go func() {
		<-sigintOrTerm
		fmt.Printf(
			"Received SIGINT/SIGTERM; initiating soft stop (try to wait for jobs to finish)\n",
		)

		softStopCtx, softStopCtxCancel := context.WithTimeout(ctx, 10*time.Second)
		defer softStopCtxCancel()

		go func() {
			select {
			case <-sigintOrTerm:
				fmt.Printf(
					"Received SIGINT/SIGTERM again; initiating hard stop (cancel everything)\n",
				)
				softStopCtxCancel()
			case <-softStopCtx.Done():
				fmt.Printf("Soft stop timeout; initiating hard stop (cancel everything)\n")
			}
		}()

		err := riverClient.Stop(softStopCtx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) &&
			!errors.Is(err, context.Canceled) {
			panic(err)
		}
		if err == nil {
			fmt.Printf("Soft stop succeeded\n")
			return
		}

		hardStopCtx, hardStopCtxCancel := context.WithTimeout(ctx, 10*time.Second)
		defer hardStopCtxCancel()

		// As long as all jobs respect context cancellation, StopAndCancel will
		// always work. However, in the case of a bug where a job blocks despite
		// being cancelled, it may be necessary to either ignore River's stop
		// result (what's shown here) or have a supervisor kill the process.
		err = riverClient.StopAndCancel(hardStopCtx)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Hard stop timeout; ignoring stop procedure and exiting unsafely\n")
		} else if err != nil {
			panic(err)
		}

		// hard stop succeeded
	}()

	// Make sure our job starts being worked before doing anything else.
	<-jobStarted

	// Cheat a little by sending a SIGTERM manually for the purpose of this
	// example (normally this will be sent by user or supervisory process). The
	// first SIGTERM tries a soft stop in which jobs are given a chance to
	// finish up.
	sigintOrTerm <- syscall.SIGTERM

	// The soft stop will never work in this example because our job only
	// respects context cancellation, but wait a short amount of time to give it
	// a chance. After it elapses, send another SIGTERM to initiate a hard stop.
	select {
	case <-riverClient.Stopped():
		// Will never be reached in this example because our job will only ever
		// finish on context cancellation.
		fmt.Printf("Soft stop succeeded\n")

	case <-time.After(100 * time.Millisecond):
		sigintOrTerm <- syscall.SIGTERM
		<-riverClient.Stopped()
	}
}
