package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Executor interface {
	Process(ctx context.Context, msg []byte) error
	Name() string
}

type RepeatableExecutor interface {
	Executor
	GenerateJob() (RepeatableJob, error)
	RescheduleJob(now time.Time) (time.Time, error)
}

type RepeatableJob struct {
	Data         Job
	ScheduledFor time.Time
}

type Job struct {
	ID            uuid.UUID
	Instructions  []byte
	FailedAttemps int32
	executor      string
}

func newJob(instructions []byte, jobExecutor string) Job {
	return Job{
		ID:           uuid.New(),
		Instructions: instructions,
		executor:     jobExecutor,
	}
}

func (j *Job) GetExecutor() string {
	return j.executor
}

func (j *Job) SetExecutor(name string) {
	j.executor = name
}
