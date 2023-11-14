package queue

import (
	"context"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/robfig/cron/v3"
)

const removeUsersExecutorName string = "remove_users_executor"

type storage interface {
	RemoveInactiveUsers(ctx context.Context, twoWeeksAgo time.Time) error
}

type removeUsersExecutor struct {
	storage  storage
	name     string
	schedule cron.Schedule
}

// GenerateJob implements RepeatableExecutor.
func (r *removeUsersExecutor) generateJob() (RepeatableJob, error) {
	nextSchedule := r.schedule.Next(time.Now())
	telemetry.Logger.Info("scheduling next job", "scheduled_for", nextSchedule)

	repeatableJob, err := newRepeatableJob(nextSchedule, JobInstructions{
		executor: removeUsersExecutorName,
	})
	if err != nil {
		return RepeatableJob{}, err
	}

	return *repeatableJob, nil
}

// Name implements RepeatableExecutor.
func (r *removeUsersExecutor) Name() string {
	return r.name
}

// RescheduleJob implements RepeatableExecutor.
func (r *removeUsersExecutor) nextRun(t time.Time) time.Time {
	return r.schedule.Next(t)
}

// process implements RepeatableExecutor.
func (r *removeUsersExecutor) process(ctx context.Context, _ []byte) error {
	t := time.Now().Add(-336 * time.Hour)
	return r.storage.RemoveInactiveUsers(ctx, t)
}

func NewRemoveUsersExecutor(cron string, storage storage) *removeUsersExecutor {
	sched, err := parser.Parse(cron)
	if err != nil {
		panic(err)
	}

	return &removeUsersExecutor{
		storage,
		removeUsersExecutorName,
		sched,
	}
}

var _ RepeatableExecutor = (*removeUsersExecutor)(nil)
