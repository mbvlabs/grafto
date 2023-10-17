package job

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Processor interface {
	Process(ctx context.Context, msg []byte) error
	Name() string
}

type SchedulingConfiguration struct {
	RepeatEvery time.Duration
	RunAt       time.Time
}

type Job struct {
	ID            uuid.UUID
	Instructions  []byte
	FailedAttemps int32
	processor     string
}

func (j *Job) GetProcessor() string {
	return j.processor
}

func (j *Job) SetProcessor(name string) {
	j.processor = name
}

const (
	StateRunning = 1
	StateFailed  = 2
	StateQueued  = 3
)
