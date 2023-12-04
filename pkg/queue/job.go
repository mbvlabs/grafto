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

type jobCreator interface {
	Schedule(t time.Time) error
	build() *job
}

type RepeatableExecutor interface {
	Executor
	generateJob() (repeatableJob, error)
	nextRun(t time.Time) time.Time
}

type repeatableJob (job)

func newRepeatableJob(scheduledFor time.Time, instructions jobInstructions) (*repeatableJob, error) {
	if instructions.executor == "" {
		return nil, ErrExecutorNotSet
	}

	if isPast := time.Now().Before(scheduledFor); !isPast {
		return nil, ErrPastTime
	}

	return &repeatableJob{
		id:           uuid.New(),
		instructions: instructions.instructions,
		executor:     instructions.executor,
		scheduledFor: scheduledFor,
	}, nil
}

type job struct {
	id            uuid.UUID
	instructions  []byte
	executor      string
	failedAttemps int32
	scheduledFor  time.Time
	isRepeating   bool
}

type jobInstructions struct {
	instructions []byte
	executor     string
}

func newJob(instructions jobInstructions) (*job, error) {
	if len(instructions.instructions) == 0 {
		return nil, ErrInstructionsNotSet
	}

	if instructions.executor == "" {
		return nil, ErrExecutorNotSet
	}

	job := &job{
		id:           uuid.New(),
		instructions: instructions.instructions,
		executor:     instructions.executor,
		scheduledFor: time.Now().Add(1500 * time.Millisecond),
	}

	return job, nil
}
