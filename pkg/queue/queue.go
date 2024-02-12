package queue

import (
	"context"
	"log/slog"
	"time"

	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

/*
ClientCfg is a thin wrapper around river.Config that provides a couple of defaults. It increases JobTimeout to 5 minutes, and uses the logger from telemetry.SetupLogger. It also sets the default queue to have a maximum of 100 workers.
*/
type ClientCfg struct {
	errorHandler      river.ErrorHandler
	fetchCooldown     time.Duration
	fetchPollInterval time.Duration
	jobTimeout        time.Duration
	logger            *slog.Logger
	periodicJobs      []*river.PeriodicJob
	queues            map[string]river.QueueConfig
	workers           *river.Workers
}

type ClientCfgOpts func(cfg *ClientCfg)

func WithErrorHandler(handler river.ErrorHandler) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.errorHandler = handler
	}
}

func WithFetchCooldown(cooldown time.Duration) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.fetchCooldown = cooldown
	}
}

func WithFetchPollInterval(interval time.Duration) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.fetchPollInterval = interval
	}
}

func WithJobTimeout(timeout time.Duration) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.jobTimeout = timeout
	}
}

func WithLogger(logger *slog.Logger) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.logger = logger
	}
}

func WithPeriodicJobs(jobs []*river.PeriodicJob) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.periodicJobs = jobs
	}
}

func WithQueues(queues map[string]river.QueueConfig) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.queues = queues
	}
}

func WithWorkers(workers *river.Workers) ClientCfgOpts {
	return func(cfg *ClientCfg) {
		cfg.workers = workers
	}
}

func NewClient(pool *pgxpool.Pool, opts ...ClientCfgOpts) *river.Client[pgx.Tx] {
	cfg := &ClientCfg{
		errorHandler:      nil,
		fetchCooldown:     100 * time.Millisecond,
		fetchPollInterval: 1 * time.Second,
		jobTimeout:        5 * time.Minute,
		logger:            telemetry.SetupLogger(),
		queues:            map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 100}},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.workers == nil {
		panic("no workers provided; queue will not process jobs")
	}

	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		ErrorHandler:      cfg.errorHandler,
		FetchCooldown:     cfg.fetchCooldown,
		FetchPollInterval: cfg.fetchCooldown,
		JobTimeout:        cfg.jobTimeout,
		Logger:            cfg.logger,
		PeriodicJobs:      cfg.periodicJobs,
		Queues:            cfg.queues,
		Workers:           cfg.workers,
	})
	if err != nil {
		panic(err)
	}

	return riverClient
}

type MailErrorHandler struct {
	logger     *slog.Logger
	mailClient *mail.Mail
	to         string
	from       string
}

func NewMailErrorHandler(logger *slog.Logger, mailClient *mail.Mail, baseSenderSignature, receiverMail string) *MailErrorHandler {
	return &MailErrorHandler{
		logger:     logger,
		from:       baseSenderSignature,
		to:         receiverMail,
		mailClient: mailClient,
	}
}

// HandleError implements river.ErrorHandler.
func (m *MailErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	m.logger.Error("error handling job", "error", err, "job_kind", job.Kind)

	if err := m.mailClient.Send(ctx, m.to, m.from, "Error handling job", "job_error", mail.FailedJob{
		ID:    job.ID,
		Kind:  job.Kind,
		Error: err.Error(),
	}); err != nil {
		m.logger.Error("error sending mail", "error", err)
		return &river.ErrorHandlerResult{}
	}

	return &river.ErrorHandlerResult{}
}

// HandlePanic implements river.ErrorHandler.
func (m *MailErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any) *river.ErrorHandlerResult {
	m.logger.Error("panic handling job", "panic", panicVal, "job_kind", job.Kind)

	if err := m.mailClient.Send(ctx, m.to, m.from, "Error handling job", "job_error", mail.FailedJob{
		ID:    job.ID,
		Kind:  job.Kind,
		Error: "panic: " + panicVal.(string),
	}); err != nil {
		m.logger.Error("error sending mail", "error", err)
		return &river.ErrorHandlerResult{}
	}

	return &river.ErrorHandlerResult{}
}

var _ river.ErrorHandler = (*MailErrorHandler)(nil)
