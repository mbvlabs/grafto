package queue

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	ErrInstructionsNotSet = errors.New("instructions not set")
	ErrExecutorNotSet     = errors.New("executor not set")
	ErrPastTime           = errors.New("cannot scheduled a job in the past")
)

type Executor interface {
	process(ctx context.Context, msg []byte) error
	Name() string
}

type RepeatableExecutor interface {
	Executor
	generateJob() (RepeatableJob, error)
	nextRun(t time.Time) time.Time
}

type RepeatableJob (Job)

func newRepeatableJob(scheduledFor time.Time, instructions JobInstructions) (*RepeatableJob, error) {
	if instructions.executor == "" {
		return nil, ErrExecutorNotSet
	}

	if isPast := time.Now().Before(scheduledFor); !isPast {
		return nil, ErrPastTime
	}

	return &RepeatableJob{
		id:           uuid.New(),
		instructions: instructions.instructions,
		executor:     instructions.executor,
		scheduledFor: scheduledFor,
	}, nil
}

type Job struct {
	id            uuid.UUID
	instructions  []byte
	executor      string
	failedAttemps int32
	scheduledFor  time.Time
	isRepeating   bool
}

func (j *Job) Schedule(t time.Time) *Job {
	j.scheduledFor = t

	return j
}

type JobInstructions struct {
	instructions []byte
	executor     string
}

func newJob(instructions JobInstructions) (*Job, error) {
	if len(instructions.instructions) == 0 {
		return nil, ErrInstructionsNotSet
	}

	if instructions.executor == "" {
		return nil, ErrExecutorNotSet
	}

	job := &Job{
		id:           uuid.New(),
		instructions: instructions.instructions,
		executor:     instructions.executor,
		scheduledFor: time.Now().Add(1500 * time.Millisecond),
	}

	return job, nil
}
