package main

import (
	"context"
	"os"

	"github.com/MBvisti/grafto/pkg/job"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/jackc/pgx/v4/pgxpool"
)

func worker(ctx context.Context, errChan chan error, id int, q *queue.Queue) {
	telemetry.Logger.Info("starting queue", "number", id)
	errChan <- q.Watch(ctx, id)
}

func main() {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	config.MaxConns = 3

	// Create a connection pool
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	// Create a channel to receive errors from goroutine
	errCh := make(chan error)

	telemetry.Logger.Info("starting queue")
	for i := 1; i <= 2; i++ {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			panic(err)
		}

		db := database.New(conn)

		q := queue.NewQueue(db)
		emailJobProcessor := job.NewEmailJobProcessor(&mailClient)
		q.RegisterHandler(emailJobProcessor)

		go worker(ctx, errCh, i, q)
	}

	err = <-errCh
	if err != nil {
		telemetry.Logger.Error("error in queue", "error", err)
	}

	select {}
}
