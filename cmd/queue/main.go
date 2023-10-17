package main

import (
	"context"
	"os"
	"time"

	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/pkg/job"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/queue"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
)

func main() {
	ctx := context.Background()

	conn := database.SetupDatabaseConnection(config.GetDatabaseURL())
	db := database.New(conn)

	q := queue.NewQueue(db)

	postmark := mail.NewPostmark(os.Getenv("POSTMARK_API_TOKEN"))
	mailClient := mail.NewMail(&postmark)

	emailJobProcessor := job.NewEmailJobProcessor(&mailClient)
	q.RegisterHandler(emailJobProcessor)

	telemetry.Logger.Info("starting queue")
	// go func() {
	// 	t := time.Tick(125 * time.Millisecond)
	// 	for {
	// 		select {
	// 		case <-t:
	// 			if err := q.StartScheduler(ctx); err != nil {
	// 				panic(err)
	// 			}
	// 		case <-ctx.Done():
	// 			telemetry.Logger.Info("shutting down queue")
	// 			return
	// 		}
	// 	}
	// }()

	newEmailJob, _ := job.CreateEmailJobMsg(job.EmailJobMsg{
		To:       "test",
		From:     "test",
		HtmlTmpl: "test",
		TextTmpl: "test",
	})

	if err := q.Push(ctx, newEmailJob, queue.SchedulingConfiguration{
		RepeatEvery: 0,
		RunAt:       time.Now().Add(30 * time.Second),
	}); err != nil {
		telemetry.Logger.Info("failed to push job", "err", err)
		panic(err)
	}

	go q.StartScheduler(ctx)

	select {}
}
