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
clientCfg is a thin wrapper around river.Config that provides a couple of defaults. It increases JobTimeout to 5 minutes, and uses the logger from telemetry.SetupLogger. It also sets the default queue to have a maximum of 100 workers.
*/
type clientCfg struct {
	errorHandler      river.ErrorHandler
	fetchCooldown     time.Duration
	fetchPollInterval time.Duration
	jobTimeout        time.Duration
	logger            *slog.Logger
	periodicJobs      []*river.PeriodicJob
	queues            *map[string]river.QueueConfig
	workers           *river.Workers
}

type ClientCfgOpts func(cfg *clientCfg)

func WithErrorHandler(handler river.ErrorHandler) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.errorHandler = handler
	}
}

func WithFetchCooldown(cooldown time.Duration) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.fetchCooldown = cooldown
	}
}

func WithFetchPollInterval(interval time.Duration) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.fetchPollInterval = interval
	}
}

func WithJobTimeout(timeout time.Duration) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.jobTimeout = timeout
	}
}

func WithLogger(logger *slog.Logger) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.logger = logger
	}
}

func WithPeriodicJobs(jobs []*river.PeriodicJob) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.periodicJobs = jobs
	}
}

func WithQueues(queues map[string]river.QueueConfig) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.queues = &queues
	}
}

func WithWorkers(workers *river.Workers) ClientCfgOpts {
	return func(cfg *clientCfg) {
		cfg.workers = workers
	}
}

/*
NewClient creates a new river.Client. It uses the provided pool to connect to the database. It uses some defaults for error handling, fetch cooldown, fetch poll interval, job timeout, and logger. For a 'read only'client, omit providing a queue.
*/
func NewClient(pool *pgxpool.Pool, opts ...ClientCfgOpts) *river.Client[pgx.Tx] {
	cfg := &clientCfg{
		fetchCooldown:     100 * time.Millisecond,
		fetchPollInterval: 1 * time.Second,
		jobTimeout:        5 * time.Minute,
		logger:            telemetry.SetupLogger(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	riverCfg := &river.Config{
		ErrorHandler:      cfg.errorHandler,
		FetchCooldown:     cfg.fetchCooldown,
		FetchPollInterval: cfg.fetchCooldown,
		JobTimeout:        cfg.jobTimeout,
		Logger:            cfg.logger,
		PeriodicJobs:      cfg.periodicJobs,
	}

	if cfg.queues != nil {
		riverCfg.Queues = *cfg.queues
		riverCfg.Workers = cfg.workers
	}

	riverClient, err := river.NewClient(riverpgxv5.New(pool), riverCfg)
	if err != nil {
		panic(err)
	}

	return riverClient
}

// MailErrorHandler is an implementation of river.ErrorHandler that sends an email when an error occurs.
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
