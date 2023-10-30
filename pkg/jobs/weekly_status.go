package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

const weeklyReportExecutor = "weekly_report_executor"

type storage interface {
	QueryUsers(ctx context.Context) ([]database.User, error)
}

type WeeklyReportExecutor struct {
	schedule cron.Schedule
	name     string
	client   emailClient
	storage  storage
}

func NewWeeklyReportExecutor(cronTab string, client emailClient, storage storage) *WeeklyReportExecutor {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronTab)
	if err != nil {
		panic(err)
	}

	return &WeeklyReportExecutor{
		schedule: schedule,
		name:     weeklyReportExecutor,
		client:   client,
		storage:  storage,
	}
}

type weeklyReportInstructions struct {
	To string `json:"to"`
}

// GenerateJob implements RepeatableExecutor.
func (q *WeeklyReportExecutor) GenerateJob() (RepeatableJob, error) {
	nextSchedule := q.schedule.Next(time.Now())

	telemetry.Logger.Info("scheduling next job", "scheduled_for", nextSchedule)

	marshaledInstructions, err := json.Marshal(weeklyReportInstructions{
		To: "mbv1406@gmail.com",
	})
	if err != nil {
		return RepeatableJob{}, err
	}

	return RepeatableJob{
		Data: Job{
			ID:           uuid.New(),
			Instructions: marshaledInstructions,
			executor:     weeklyReportExecutor,
		},
		ScheduledFor: nextSchedule,
	}, nil
}

// Name implements RepeatableExecutor.
func (w *WeeklyReportExecutor) Name() string {
	return w.name
}

// Process implements RepeatableExecutor.
func (w *WeeklyReportExecutor) Process(ctx context.Context, msg []byte) error {
	var instructions weeklyReportInstructions
	if err := json.Unmarshal(msg, &instructions); err != nil {
		return err
	}

	// get users
	// users, err := w.storage.QueryUsers(ctx)
	// if err != nil {
	// 	return err
	// }

	// send email
	return w.client.Send(ctx, instructions.To, "newsletter@mortenvistisen.com", "test", "weekly_report", mail.WeeklyStatusReport{
		NewUsers: 3,
	})
}

// RescheduleJob implements RepeatableExecutor.
func (q *WeeklyReportExecutor) RescheduleJob(now time.Time) time.Time {
	return q.schedule.Next(now)
}

var _ RepeatableExecutor = (*WeeklyReportExecutor)(nil)
