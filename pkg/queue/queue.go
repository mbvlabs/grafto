package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"time"

	"github.com/MBvisti/grafto/pkg/jobs"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const DefaultMaxRetries = 5

const (
	stateRunning = 1
	stateFailed  = 2
	stateQueued  = 3
)

type queueStorage interface {
	InsertJob(ctx context.Context, arg database.InsertJobParams) error
	QueryJobs(ctx context.Context, arg database.QueryJobsParams) ([]database.Queue, error)
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	ClearQueue(ctx context.Context) error
	CheckIfRepeatableJobExists(ctx context.Context, repeatableJobID sql.NullString) (bool, error)
	RescheduleRepeatableJob(ctx context.Context, arg database.RescheduleRepeatableJobParams) error
}

type Queue struct {
	maxRetries          int32
	executors           map[string]jobs.Executor
	repeatableExecutors map[string]jobs.RepeatableExecutor
	storage             queueStorage
	hasher              hash.Hash
}

func NewQueue(storage queueStorage) *Queue {
	hasher := fnv.New64a()
	return &Queue{
		DefaultMaxRetries,
		map[string]jobs.Executor{},
		map[string]jobs.RepeatableExecutor{},
		storage,
		hasher,
	}
}

func (q *Queue) Clear(ctx context.Context) error {
	return q.storage.ClearQueue(ctx)
}

func (q *Queue) delete(ctx context.Context, id uuid.UUID) error {
	return q.storage.DeleteJob(ctx, id)
}

func (q *Queue) fail(ctx context.Context, j jobs.Job) error {
	params := database.FailJobParams{
		UpdatedAt: time.Now(),
		ID:        j.ID,
	}

	if j.FailedAttemps == q.maxRetries {
		params.State = stateFailed
	} else {
		params.State = stateQueued
		params.ScheduledFor = time.Now().Add(1500 * time.Millisecond)
	}

	return q.storage.FailJob(ctx, params)
}

func (q *Queue) pull(ctx context.Context, pullTime time.Time) ([]database.Queue, error) {
	return q.storage.QueryJobs(ctx, database.QueryJobsParams{
		State:               stateRunning,
		UpdatedAt:           time.Now(),
		Limit:               10,
		InnerState:          stateQueued,
		InnerScheduledFor:   time.Now(),
		InnerFailedAttempts: int32(q.maxRetries),
	})
}

func (q *Queue) Push(ctx context.Context, jobPayload jobs.Job) error {
	msg := pgtype.JSONB{}
	if err := msg.Set(jobPayload.Instructions); err != nil {
		telemetry.Logger.Error("failed to set job instructions", "error", err)
		return err
	}

	t := time.Now()
	return q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:           jobPayload.ID,
		CreatedAt:    t,
		UpdatedAt:    t,
		State:        stateQueued,
		Message:      msg,
		Processor:    jobPayload.GetExecutor(),
		ScheduledFor: t.Add(1500 * time.Millisecond),
	})
}

func (q *Queue) RegisterExecutors(executors []jobs.Executor) {
	for _, executor := range executors {
		q.executors[executor.Name()] = executor
	}
}

func (q *Queue) RegisterRepeatingExecutors(ctx context.Context, repeatExecutors []jobs.RepeatableExecutor) error {
	for _, executor := range repeatExecutors {
		q.repeatableExecutors[executor.Name()] = executor

		job, err := executor.GenerateJob()
		if err != nil {
			return err
		}

		marshaledJob, err := json.Marshal(job.Data.Instructions)
		if err != nil {
			return err
		}

		q.hasher.Write(marshaledJob)
		repeatJobID := fmt.Sprintf("%x", q.hasher.Sum(nil))

		if exists, err := q.storage.CheckIfRepeatableJobExists(
			ctx, sql.NullString{String: repeatJobID, Valid: true}); err != nil {
			return err
		} else if exists {
			telemetry.Logger.Info("repeatable job already exists, skipping", "job", repeatJobID)
			return nil
		}

		msg := pgtype.JSONB{}
		if err := msg.Set(job.Data.Instructions); err != nil {
			telemetry.Logger.Error("failed to set job instructions", "error", err)
			return err
		}

		err = q.storage.InsertJob(ctx, database.InsertJobParams{
			ID:              uuid.New(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			ScheduledFor:    job.ScheduledFor,
			State:           stateQueued,
			Message:         msg,
			Processor:       job.Data.GetExecutor(),
			RepeatableJobID: sql.NullString{String: repeatJobID, Valid: true},
		})
		if err != nil {
			telemetry.Logger.Error("failed to insert job", "error", err)
			return err
		}
	}

	return nil
}

func (q *Queue) Watch(ctx context.Context) error {
	for {
		t := time.Now()
		queuedJobs, err := q.pull(ctx, t)
		if err != nil {
			return err
		}

		telemetry.Logger.Info("pulled jobs", "jobs", 1 < len(queuedJobs))

		for _, queuedJob := range queuedJobs {
			isRepeating := queuedJob.RepeatableJobID.Valid
			job := jobs.Job{
				ID:            queuedJob.ID,
				Instructions:  queuedJob.Message.Bytes,
				FailedAttemps: queuedJob.FailedAttempts,
			}
			job.SetExecutor(queuedJob.Processor)

			err := q.handleJob(ctx, job, isRepeating)
			if err != nil {
				telemetry.Logger.Error("failed to handle job", "error", err)

				if err := q.storage.FailJob(ctx, database.FailJobParams{
					State:     stateFailed,
					UpdatedAt: time.Now(),
					ID:        job.ID,
				}); err != nil {
					return err
				}
			}
		}

		time.Sleep(125 * time.Millisecond)
	}
}

func (q *Queue) handleJob(ctx context.Context, j jobs.Job, isRepeating bool) error {
	if !isRepeating {
		if err := q.executors[j.GetExecutor()].Process(ctx, j.Instructions); err != nil {
			telemetry.Logger.Error("failed to process job", "error", err)
			return q.fail(ctx, j)
		}
	}

	if isRepeating {
		if err := q.repeatableExecutors[j.GetExecutor()].Process(ctx, j.Instructions); err != nil {
			telemetry.Logger.Error("failed to process job", "error", err)
			return q.fail(ctx, j)
		}

		scheduledFor, err := q.repeatableExecutors[j.GetExecutor()].RescheduleJob()
		if err != nil {
			return err
		}

		return q.storage.RescheduleRepeatableJob(ctx, database.RescheduleRepeatableJobParams{
			State:        stateQueued,
			UpdatedAt:    time.Now(),
			ScheduledFor: scheduledFor,
			ID:           j.ID,
		})
	}

	return q.delete(ctx, j.ID)
}
