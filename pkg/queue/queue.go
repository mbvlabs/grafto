package queue

import (
	"context"
	"database/sql"
	"fmt"
	"hash"
	"hash/fnv"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const (
	stateRunning = 1
	stateFailed  = 2
	stateQueued  = 3
	maxRetries   = 5
)

type queueStorage interface {
	QueryJobs(ctx context.Context, params database.QueryJobsParams) ([]database.Job, error)
	InsertJob(ctx context.Context, params database.InsertJobParams) error
	RepeatableJobExists(ctx context.Context, repeatableID sql.NullString) (bool, error)
}

type Queue struct {
	storage    queueStorage
	maxRetries int
	hasher     hash.Hash
}

func New(storage queueStorage) *Queue {
	hasher := fnv.New64a()

	return &Queue{
		storage,
		maxRetries,
		hasher,
	}
}

func (q *Queue) pull(ctx context.Context) ([]database.Job, error) {
	return q.storage.QueryJobs(ctx, database.QueryJobsParams{
		State:               stateRunning,
		UpdatedAt:           time.Now(),
		Limit:               50,
		InnerState:          stateQueued,
		InnerScheduledFor:   time.Now(),
		InnerFailedAttempts: int32(q.maxRetries),
	})
}

func (q *Queue) Push(ctx context.Context, payload *Job) error {
	if err := payload.validate(); err != nil {
		return err
	}

	var instructions pgtype.JSONB
	if err := instructions.Set(payload.instructions); err != nil {
		return err
	}

	return q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		State:        stateQueued,
		Instructions: instructions,
		Executor:     payload.executor,
		ScheduledFor: payload.scheduledFor,
	})
}

func (q *Queue) InitilizeRepeatingJobs(ctx context.Context, executors map[string]RepeatableExecutor) error {
	for name, executor := range executors {
		job, err := executor.generateJob()
		if err != nil {
			return err
		}

		var instructions pgtype.JSONB
		if err := instructions.Set(job.instructions); err != nil {
			return err
		}

		q.hasher.Write(job.instructions)
		repeatJobID := fmt.Sprintf("%x", q.hasher.Sum(nil))

		if exists, err := q.storage.RepeatableJobExists(
			ctx, sql.NullString{String: repeatJobID, Valid: true}); err != nil {
			return err
		} else if exists {
			telemetry.Logger.Info("repeatable job already exists, skipping", "job", repeatJobID)
			return nil
		}

		err = q.storage.InsertJob(ctx, database.InsertJobParams{
			ID:           uuid.New(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ScheduledFor: job.scheduledFor,
			State:        stateQueued,
			Instructions: instructions,
			Executor:     name,
			RepeatableID: sql.NullString{String: repeatJobID, Valid: true},
		})
		if err != nil {
			telemetry.Logger.Error("failed to insert job", "error", err)
			return err
		}
	}

	return nil
}

func (q *Queue) Start(ctx context.Context, jobs chan<- []Job) error {
	telemetry.Logger.Info("starting to watch the queue")

	for {
		queuedJobs, err := q.pull(ctx)
		if err != nil {
			return err
		}

		j := make([]Job, 0, len(queuedJobs))
		for _, queuedJob := range queuedJobs {
			j = append(j, Job{
				id:            queuedJob.ID,
				instructions:  queuedJob.Instructions.Bytes,
				executor:      queuedJob.Executor,
				failedAttemps: queuedJob.FailedAttempts,
			})
		}
		jobs <- j

		time.Sleep(125 * time.Millisecond)
	}
}
