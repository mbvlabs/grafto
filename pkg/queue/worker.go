package queue

import (
	"context"
	"time"

	"github.com/MBvisti/grafto/pkg/jobs"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type workerStorage interface {
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	RescheduleRepeatableJob(ctx context.Context, arg database.RescheduleRepeatableJobParams) error
}

type Worker struct {
	ID                  uuid.UUID
	jobsChan            <-chan []database.Job
	storage             workerStorage
	executors           map[string]jobs.Executor
	repeatableExecutors map[string]jobs.RepeatableExecutor
}

func NewWorker(
	id uuid.UUID, jobs <-chan []database.Job, storage workerStorage, executors map[string]jobs.Executor,
	repeatableExecutors map[string]jobs.RepeatableExecutor) *Worker {
	return &Worker{
		ID:                  id,
		jobsChan:            jobs,
		storage:             storage,
		executors:           executors,
		repeatableExecutors: repeatableExecutors,
	}
}

var Counter = 0

func (w *Worker) Handle() {
	// semaphore := make(chan uuiiiid.UUID, 5)

	for {
		// jobs := <-w.jobsChan

		for _, job := range <-w.jobsChan {
			// semaphore <- job.ID

			if job.RepeatableJobID.Valid {
				// time.Sleep(500 * time.Millisecond)
				if err := w.processRepeatable(job); err != nil {
					telemetry.Logger.Error("failed to process job", "job", job.ID, "error", err)
					continue
				}
				// <-semaphore
			}

			if !job.RepeatableJobID.Valid {
				if err := w.processOneOff(job); err != nil {
					telemetry.Logger.Error("failed to process job", "job", job.ID, "error", err)
					continue
				}
				// <-semaphore
			}
		}
	}
}

func (w *Worker) processRepeatable(job database.Job) error {
	t := time.Now()
	telemetry.Logger.Info("call time", "now", t)

	executor := w.repeatableExecutors[job.Executor]
	err := executor.Process(context.Background(), job.Instructions)
	if err != nil {
		telemetry.Logger.Error("failed to process job", "job", job.ID, "error", err)
		return w.failJob(context.Background(), job.ID, job.FailedAttempts)
	}

	scheduledFor, err := executor.RescheduleJob(t)
	if err != nil {
		return w.failJob(context.Background(), job.ID, job.FailedAttempts)
	}

	if err := w.storage.RescheduleRepeatableJob(context.Background(), database.RescheduleRepeatableJobParams{
		State:        stateQueued,
		UpdatedAt:    database.ConvertTime(time.Now()),
		ScheduledFor: database.ConvertTime(scheduledFor),
		ID:           job.ID,
	}); err != nil {
		return w.failJob(context.Background(), job.ID, job.FailedAttempts)
	}

	return nil
}

func (w *Worker) processOneOff(job database.Job) error {
	executor := w.executors[job.Executor]
	if err := executor.Process(context.Background(), job.Instructions); err != nil {
		return w.failJob(context.Background(), job.ID, job.FailedAttempts)
	}

	return w.storage.DeleteJob(context.Background(), job.ID)
}

func (w *Worker) failJob(ctx context.Context, id uuid.UUID, failedAttemps int32) error {
	params := database.FailJobParams{
		UpdatedAt: database.ConvertTime(time.Now()),
		ID:        id,
	}

	if failedAttemps == DefaultMaxRetries {
		params.State = stateFailed
	} else {
		params.State = stateQueued
		params.ScheduledFor = database.ConvertTime(time.Now().Add(1500 * time.Millisecond))
	}

	return w.storage.FailJob(context.Background(), params)
}
