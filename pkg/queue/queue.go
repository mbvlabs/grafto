package queue

import (
	"context"
	"time"

	"github.com/MBvisti/grafto/pkg/job"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const DefaultMaxRetries = 5

type queueStorage interface {
	InsertJob(ctx context.Context, arg database.InsertJobParams) error
	QueryJobs(ctx context.Context, arg database.QueryJobsParams) ([]database.Queue, error)
	FailJob(ctx context.Context, arg database.FailJobParams) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	ClearQueue(ctx context.Context) error
}

type Queue struct {
	maxRetries    int32
	jobProcessors map[string]job.Processor
	storage       queueStorage
}

func NewQueue(storage queueStorage) *Queue {
	return &Queue{
		DefaultMaxRetries,
		map[string]job.Processor{},
		storage,
	}
}

// Clear implements Queuer.
func (*Queue) Clear(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteTask implements Queuer.
func (q *Queue) delete(ctx context.Context, id uuid.UUID) error {
	return q.storage.DeleteJob(ctx, id)
}

// FailTask implements Queuer.
func (q *Queue) fail(ctx context.Context, j job.Job) error {
	params := database.FailJobParams{
		UpdatedAt: time.Now(),
		ID:        j.ID,
	}

	if j.FailedAttemps == q.maxRetries {
		params.State = job.StateFailed
	} else {
		params.State = job.StateQueued
		params.ScheduledFor = time.Now().Add(10 * time.Second)
	}

	return q.storage.FailJob(ctx, params)
}

// Pull implements Queuer.
func (q *Queue) pull(ctx context.Context) ([]job.Job, error) {
	queuedJobs, err := q.storage.QueryJobs(ctx, database.QueryJobsParams{
		State:               job.StateRunning,
		UpdatedAt:           time.Now(),
		Limit:               50,
		InnerState:          job.StateQueued,
		InnerScheduledFor:   time.Now(),
		InnerFailedAttempts: int32(q.maxRetries),
	})
	if err != nil {
		telemetry.Logger.Error("failed to query tasks", "error", err)
		return nil, err
	}

	var jobs []job.Job
	for _, queuedJob := range queuedJobs {
		job := job.Job{
			ID:            queuedJob.ID,
			Instructions:  queuedJob.Message.Bytes,
			FailedAttemps: queuedJob.FailedAttempts,
		}
		job.SetProcessor(queuedJob.Processor)

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Push implements Queuer.
func (q *Queue) Push(ctx context.Context, j job.Job, config SchedulingConfiguration) error {
	// id, err := uuid.Parse(ulid.Make().String()) // TODO: handle error
	// if err != nil {
	// 	telemetry.Logger.Error("failed to parse uuid", "error", err)
	// 	return err
	// }

	msg := pgtype.JSONB{}
	if err := msg.Set(j.Instructions); err != nil {
		return err
	}

	return q.storage.InsertJob(ctx, database.InsertJobParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ScheduledFor: config.RunAt,
		State:        job.StateQueued,
		Message:      msg,
		Processor:    j.GetProcessor(),
	})
}

func (q *Queue) RegisterHandler(processor job.Processor) {
	q.jobProcessors[processor.Name()] = processor
}

func (q *Queue) StartScheduler(ctx context.Context) {
	for {
		telemetry.Logger.Info("pulling jobs")
		jobs, err := q.pull(ctx)
		if err != nil {
			panic(err)
		}

		for _, job := range jobs {
			err := q.handleJob(ctx, job)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(125 * time.Millisecond)

	}
}

type SchedulingConfiguration struct {
	RepeatEvery time.Duration
	RunAt       time.Time
}

func (q *Queue) handleJob(ctx context.Context, job job.Job) error {
	processor := q.jobProcessors[job.GetProcessor()]

	if err := processor.Process(ctx, job.Instructions); err != nil {
		telemetry.Logger.Error("failed to process job", "error", err)
		return q.fail(ctx, job)
	}

	if err := q.delete(ctx, job.ID); err != nil {
		return err
	}

	return nil
}
