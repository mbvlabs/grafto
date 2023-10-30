package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/MBvisti/grafto/pkg/jobs"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

const DefaultMaxRetries = 5

const (
	stateRunning = 1
	stateFailed  = 2
	stateQueued  = 3
)

type JobPayload struct {
	Instructions []byte `json:"instructions"`
	Executor     string `json:"executor"`
}

type queueStorage interface {
	InsertJob(ctx context.Context, arg database.InsertJobParams) error
	QueryJobs(ctx context.Context, arg database.QueryJobsParams) ([]database.Job, error)
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	ClearJobs(ctx context.Context) error
	CheckIfRepeatableJobExists(ctx context.Context, repeatableJobID sql.NullString) (bool, error)
	RescheduleRepeatableJob(ctx context.Context, arg database.RescheduleRepeatableJobParams) error
}

type Queue struct {
	maxRetries int32
	storage    queueStorage
	hasher     hash.Hash
}

func NewQueue(storage queueStorage) *Queue {
	hasher := fnv.New64a()
	return &Queue{
		DefaultMaxRetries,
		storage,
		hasher,
	}
}

func (q *Queue) Clear(ctx context.Context) error {
	return q.storage.ClearJobs(ctx)
}

func (q *Queue) pull(ctx context.Context, pullTime time.Time) ([]database.Job, error) {
	return q.storage.QueryJobs(ctx, database.QueryJobsParams{
		State:               stateRunning,
		UpdatedAt:           database.ConvertTime(pullTime),
		Limit:               50,
		InnerState:          stateQueued,
		InnerScheduledFor:   database.ConvertTime(pullTime),
		InnerFailedAttempts: int32(q.maxRetries),
	})
}

func (q *Queue) Push(ctx context.Context, payload JobPayload) error {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 0 and (4000 - 2000) = 2000
	randomNumber := rand.Intn(2001)

	// Add the minimum value (2000) to the random number to get a value between 2000 and 4000
	randomNumberBetween2000And4000 := 2000 + randomNumber

	t := time.Now()
	return q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:           uuid.New(),
		CreatedAt:    database.ConvertTime(t),
		UpdatedAt:    database.ConvertTime(t),
		State:        stateQueued,
		Instructions: payload.Instructions,
		Executor:     payload.Executor,
		ScheduledFor: database.ConvertTime(t.Add(time.Duration(randomNumberBetween2000And4000) * time.Millisecond)),
	})
}

func (q *Queue) InitilizeRepeatingJobs(ctx context.Context, repeatExecutor jobs.RepeatableExecutor) error {
	job, err := repeatExecutor.GenerateJob()
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

	err = q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:              uuid.New(),
		CreatedAt:       database.ConvertTime(time.Now()),
		UpdatedAt:       database.ConvertTime(time.Now()),
		ScheduledFor:    database.ConvertTime(job.ScheduledFor),
		State:           stateQueued,
		Instructions:    job.Data.Instructions,
		Executor:        job.Data.GetExecutor(),
		RepeatableJobID: sql.NullString{String: repeatJobID, Valid: true},
	})
	if err != nil {
		telemetry.Logger.Error("failed to insert job", "error", err)
		return err
	}

	return nil
}

func (q *Queue) Watch(ctx context.Context, jobs chan<- []database.Job, errCh chan<- error) {
	telemetry.Logger.Info("starting the watch")
	l := 1
	for {
		t := time.Now()
		queuedJobs, err := q.pull(ctx, t)
		if err != nil {
			errCh <- err
		} else {
			jobs <- queuedJobs
		}

		telemetry.Logger.Info("sending jobs to channel", "count_job", len(queuedJobs), "it", l)

		l++
		time.Sleep(1000 * time.Millisecond)
	}
}
