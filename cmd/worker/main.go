package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

func main() {
	ctx := context.Background()
	cfg := config.New()

	logger := telemetry.SetupLogger()

	postmark := mail.NewPostmark(cfg.ExternalProviders.PostmarkApiToken)
	mailClient := mail.NewMail(&postmark)

	queueDbPool, err := pgxpool.New(context.Background(), cfg.Db.GetQueueUrlString())
	if err != nil {
		panic(err)
	}

	if err := queueDbPool.Ping(ctx); err != nil {
		panic(err)
	}

	conn := database.SetupDatabasePool(context.Background(), cfg.Db.GetUrlString())
	db := database.New(conn)

	jobStarted := make(chan struct{})

	workers, err := queue.SetupWorkers(queue.WorkerDependencies{
		Db:         db,
		MailClient: mailClient,
	})
	if err != nil {
		panic(err)
	}

	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return queue.RemoveUnverifiedUsersJobArgs{}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}

	q := map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 100}}
	riverClient := queue.NewClient(queueDbPool, queue.WithQueues(q), queue.WithWorkers(workers), queue.WithLogger(logger), queue.WithPeriodicJobs(periodicJobs))

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
		fmt.Printf("Received SIGINT/SIGTERM; initiating soft stop (try to wait for jobs to finish)\n")

		softStopCtx, softStopCtxCancel := context.WithTimeout(ctx, 10*time.Second)
		defer softStopCtxCancel()

		go func() {
			select {
			case <-sigintOrTerm:
				fmt.Printf("Received SIGINT/SIGTERM again; initiating hard stop (cancel everything)\n")
				softStopCtxCancel()
			case <-softStopCtx.Done():
				fmt.Printf("Soft stop timeout; initiating hard stop (cancel everything)\n")
			}
		}()

		err := riverClient.Stop(softStopCtx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
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
