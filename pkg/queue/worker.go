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
	RescheduleJob(ctx context.Context, arg database.RescheduleJobParams) error
}

type Worker struct {
	jobsChan           chan []Job
	storage            workerStorage
	executors          map[string]Executor
	repeatableExecutor map[string]RepeatableExecutor
}

func NewWorker(jobsChan chan []Job, storage workerStorage, executors map[string]Executor,
	repeatableExecutor map[string]RepeatableExecutor) *Worker {
	return &Worker{
		jobsChan,
		storage,
		executors,
		repeatableExecutor,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		for _, job := range <-w.jobsChan {
			if executor, ok := w.executors[job.executor]; ok {
				if err := w.processOneOff(executor, job); err != nil {
					continue
				}
			}

			if repeatableExecutor, ok := w.repeatableExecutor[job.executor]; ok {
				telemetry.Logger.Info("processing repeatable job", "job", job)
				if err := w.processRepeatable(repeatableExecutor, job); err != nil {
					continue
				}
			}
		}
	}
}

func (w *Worker) processRepeatable(executor RepeatableExecutor, job Job) error {
	err := executor.process(context.Background(), job.instructions)
	if err != nil {
		telemetry.Logger.Error("failed to process job", "job", job.id, "error", err)
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

	return nil
}

func (w *Worker) processOneOff(executor Executor, job Job) error {
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
