package queue

import (
	"context"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type workerStorage interface {
	QueryJobs(ctx context.Context, params database.QueryJobsParams) ([]database.Job, error)
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	RescheduleJob(ctx context.Context, arg database.RescheduleJobParams) error
}

type Worker struct {
	jobsChan           chan []job
	storage            workerStorage
	executors          map[string]Executor
	repeatableExecutor map[string]RepeatableExecutor
}

func NewWorker(storage workerStorage, executors map[string]Executor,
	repeatableExecutor map[string]RepeatableExecutor) *Worker {
	jobsChan := make(chan []job)

	return &Worker{
		jobsChan,
		storage,
		executors,
		repeatableExecutor,
	}
}

func (w *Worker) WatchQueue(ctx context.Context) error {
	telemetry.Logger.Info("starting to watch the queue")

	for {
		queuedJobs, err := w.storage.QueryJobs(ctx, database.QueryJobsParams{
			State:               stateRunning,
			UpdatedAt:           time.Now(),
			Limit:               50,
			InnerState:          stateQueued,
			InnerScheduledFor:   time.Now(),
			InnerFailedAttempts: int32(maxRetries),
		})

		if err != nil {
			return err
		}

		j := make([]job, 0, len(queuedJobs))
		for _, queuedJob := range queuedJobs {
			j = append(j, job{
				id:            queuedJob.ID,
				instructions:  queuedJob.Instructions.Bytes,
				executor:      queuedJob.Executor,
				failedAttemps: queuedJob.FailedAttempts,
			})
		}
		w.jobsChan <- j

		time.Sleep(125 * time.Millisecond)
	}
}

func (w *Worker) ProcessQueue(ctx context.Context) {
	for {
		for _, job := range <-w.jobsChan {
			if executor, ok := w.executors[job.executor]; ok {
				if err := w.processOneOff(executor, job); err != nil {
					continue
				}
			}

			if repeatableExecutor, ok := w.repeatableExecutor[job.executor]; ok {
				if err := w.processRepeatable(repeatableExecutor, job); err != nil {
					continue
				}
			}
		}
	}
}

func (w *Worker) processOneOff(executor Executor, job job) error {
	telemetry.Logger.Info("starting to process one-off job", "job", job, "executor", executor.Name())
	if err := executor.process(context.Background(), job.instructions); err != nil {
		err := w.failJob(context.Background(), job.id, job.failedAttemps)
		if err != nil {
			telemetry.Logger.Error("could not fail job", "error", err, "job", job)
			return err
		}
	}

	if err := w.storage.DeleteJob(context.Background(), job.id); err != nil {
		telemetry.Logger.Error("could not delete job", "error", err, "job", job)
		return err
	}

	telemetry.Logger.Info("finished processing job", "job", job, "executor", executor.Name())
	return nil
}

func (w *Worker) processRepeatable(executor RepeatableExecutor, job job) error {
	telemetry.Logger.Info("starting to process repeatable job", "job_id", job.id, "executor", executor.Name())

	if err := executor.process(context.Background(), job.instructions); err != nil {
		telemetry.Logger.Error("failed to process job", "job_id", job.id, "error", err)
		return w.failJob(context.Background(), job.id, job.failedAttemps)
	}

	if err := w.storage.RescheduleJob(context.Background(), database.RescheduleJobParams{
		State:        stateQueued,
		UpdatedAt:    time.Now(),
		ScheduledFor: executor.nextRun(time.Now()),
		ID:           job.id,
	}); err != nil {
		return w.failJob(context.Background(), job.id, job.failedAttemps)
	}

	telemetry.Logger.Info("finished processing repeatable job", "job", job, "executor", executor.Name())
	return nil
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
