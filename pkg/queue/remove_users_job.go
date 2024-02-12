package queue

import (
	"context"
	"time"

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
