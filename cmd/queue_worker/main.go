package main

import (
	"context"
	"os"

	"github.com/MBvisti/grafto/pkg/jobs"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

const DefaultWorkerCount = 5

func initRepeatingJobs(ctx context.Context, q *queue.Queue, repeatableExecutors map[string]jobs.RepeatableExecutor) error {
	for _, v := range repeatableExecutors {
		err := q.InitilizeRepeatingJobs(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	queuedJobsStream := make(chan []database.Job)

	// Create a connection pool
	pool := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	defer pool.Close()

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	db := database.New(pool)
	q := queue.NewQueue(db)

	emailJobExecutor := jobs.NewEmailJobExecutor(&mailClient)
	executors := map[string]jobs.Executor{
		emailJobExecutor.Name(): emailJobExecutor,
	}

	weeklyReport := jobs.NewWeeklyReportExecutor("* * * * *", &mailClient, db)
	repeatableExecutors := map[string]jobs.RepeatableExecutor{weeklyReport.Name(): weeklyReport}
	if err := initRepeatingJobs(ctx, q, repeatableExecutors); err != nil {
		panic(err)
	}

	// type EmailInstructions struct {
	// 	To       string      `json:"to"`
	// 	From     string      `json:"from"`
	// 	TmplName string      `json:"tmpl_name"`
	// 	Payload  interface{} `json:"payload"`
	// }
	// go func() {
	// 	inst, _ := json.Marshal(EmailInstructions{
	// 		To:       "dsajkdlasmk",
	// 		From:     "dkasmldmsakl",
	// 		TmplName: "weekly_report",
	// 		Payload:  nil,
	// 	})
	// 	for i := 0; i < 100; i++ {
	// 		_ = q.Push(ctx, queue.JobPayload{
	// 			Instructions: inst,
	// 			Executor:     "testing",
	// 		})
	// 	}
	// }()

	worker := queue.NewWorker(uuid.New(), queuedJobsStream, db, executors, repeatableExecutors)
	go worker.Handle()

	// Create a channel to receive errors from goroutine
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			telemetry.Logger.Error("error in queue", "error", err)
		}
	}()

	q.Watch(ctx, queuedJobsStream, errCh)
}
