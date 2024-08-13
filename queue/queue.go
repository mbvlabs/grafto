package queue

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
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
NewClient creates a new river.Client. It uses the provided pool to connect to the database. It uses some defaults for error handling, fetch cooldown, fetch poll interval, job timeout, and logger. For a 'read only' client, omit the queue.
*/
func NewClient(pool *pgxpool.Pool, opts ...ClientCfgOpts) *river.Client[pgx.Tx] {
	cfg := &clientCfg{
		fetchCooldown:     100 * time.Millisecond,
		fetchPollInterval: 1 * time.Second,
		jobTimeout:        5 * time.Minute,
		logger:            slog.Default(),
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
