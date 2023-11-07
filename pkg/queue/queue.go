package queue

import (
	"context"
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
	maxRetries   = 5 // TODO: update this in the code
)

type queueStorage interface {
	QueryJobs(ctx context.Context, params database.QueryJobsParams) ([]database.Job, error)
	InsertJob(ctx context.Context, params database.InsertJobParams) error
}

type Queue struct {
	storage    queueStorage
	maxRetries int
}

func New(storage queueStorage) *Queue {
	return &Queue{
		storage,
		maxRetries,
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
		ScheduledFor: time.Now().Add(1500 * time.Millisecond),
	})
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
