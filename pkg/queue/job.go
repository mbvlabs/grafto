package queue

import (
	"context"

	"github.com/google/uuid"
)

type Executor interface {
	Process(ctx context.Context, msg []byte) error
	Name() string
}

type Job struct {
	id            uuid.UUID
	instructions  []byte
	executor      string
	failedAttemps int32
}

func createJob(instructions []byte, executor string) *Job {
	return &Job{
		id:           uuid.New(),
		instructions: instructions,
		executor:     executor,
	}
}
