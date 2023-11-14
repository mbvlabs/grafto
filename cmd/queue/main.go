package main

import (
	"context"
	"os"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
)

func main() {
	ctx := context.Background()

	queuedJobsStream := make(chan []queue.Job)

	databaseConnection := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	defer databaseConnection.Close(ctx)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	db := database.New(databaseConnection)
	q := queue.New(db)

	emailJobExecutor := queue.NewEmailExecutor(&mailClient)
	executors := map[string]queue.Executor{
		emailJobExecutor.Name(): emailJobExecutor,
	}

	repeatableExecutors := map[string]queue.RepeatableExecutor{}
	if err := q.InitilizeRepeatingJobs(ctx, repeatableExecutors); err != nil {
		panic(err)
	}

	worker := queue.NewWorker(queuedJobsStream, db, executors, repeatableExecutors)
	go worker.Start(ctx)

	if err := q.Start(ctx, queuedJobsStream); err != nil {
		telemetry.Logger.ErrorContext(ctx, "watching the queue failed", "error", err)
		panic(err)
	}
}
