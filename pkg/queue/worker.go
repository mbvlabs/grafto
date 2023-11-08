package queue

import (
	"context"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type workerStorage interface {
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
}

type Worker struct {
	jobsChan  chan []Job
	storage   workerStorage
	executors map[string]Executor
}

func NewWorker(jobsChan chan []Job, storage workerStorage, executors map[string]Executor) *Worker {
	return &Worker{
		jobsChan:  jobsChan,
		storage:   storage,
		executors: executors,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		for _, job := range <-w.jobsChan {
			executor := w.executors[job.executor]
			if err := executor.process(context.Background(), job.instructions); err != nil {
				err := w.failJob(context.Background(), job.id, job.failedAttemps)
				if err != nil {
					telemetry.Logger.Error("could not fail job", "error", err, "job", job)
					continue
				}
			}

			if err := w.storage.DeleteJob(context.Background(), job.id); err != nil {
				telemetry.Logger.Error("could not delete job", "error", err, "job", job)
				continue
			}
		}
	}
}

func (w *Worker) failJob(ctx context.Context, id uuid.UUID, failedAttemps int32) error {
	params := database.FailJobParams{
		UpdatedAt: time.Now(),
		ID:        id,
	}

	if failedAttemps == maxRetries {
		params.State = stateFailed
	} else {
		params.State = stateQueued
		params.ScheduledFor = time.Now().Add(1500 * time.Millisecond)
	}

	return w.storage.FailJob(context.Background(), params)
}
