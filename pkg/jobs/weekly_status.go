package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/adhocore/gronx"
	"github.com/google/uuid"
)

const weeklyReportExecutor = "weekly_report_executor"

type storage interface {
	QueryUsers(ctx context.Context) ([]database.User, error)
}

type WeeklyReportExecutor struct {
	schedule string
	name     string
	client   emailClient
	storage  storage
}

func NewWeeklyReportExecutor(cron string, client emailClient, storage storage) *WeeklyReportExecutor {
	// maybe check cron
	return &WeeklyReportExecutor{
		schedule: cron,
		name:     weeklyReportExecutor,
		client:   client,
		storage:  storage,
	}
}

type WeeklyReportInstructions struct {
	To string `json:"to"`
}

// GenerateJob implements RepeatableExecutor.
func (q *WeeklyReportExecutor) GenerateJob() (RepeatableJob, error) {
	nextSchedule, err := gronx.NextTick(q.schedule, true)
	if err != nil {
		return RepeatableJob{}, err
	}

	marshaledInstructions, err := json.Marshal(WeeklyReportInstructions{
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
	var instructions WeeklyReportInstructions
	if err := json.Unmarshal(msg, &instructions); err != nil {
		return err
	}

	// get users
	users, err := w.storage.QueryUsers(ctx)
	if err != nil {
		return err
	}

	telemetry.Logger.Info("sending weekly report", "user_count", len(users))

	// send email
	return nil
}

// RescheduleJob implements RepeatableExecutor.
func (q *WeeklyReportExecutor) RescheduleJob() (time.Time, error) {
	return gronx.NextTick(q.schedule, true)
}

var _ RepeatableExecutor = (*WeeklyReportExecutor)(nil)
