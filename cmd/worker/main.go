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

	databaseConnection := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	defer databaseConnection.Close(ctx)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	db := database.New(databaseConnection)

	emailJobExecutor := queue.NewEmailExecutor(&mailClient)
	executors := map[string]queue.Executor{
		emailJobExecutor.Name(): emailJobExecutor,
	}

	repeatableExecutors := map[string]queue.RepeatableExecutor{}

	worker := queue.NewWorker(db, executors, repeatableExecutors)
	go worker.Process(ctx)

	if err := worker.WatchQueue(ctx); err != nil {
		telemetry.Logger.ErrorContext(ctx, "watching the queue failed", "error", err)
		panic(err)
	}
}
