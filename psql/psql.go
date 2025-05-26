package psql

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/exaring/otelpgx"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/psql/queue"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"github.com/riverqueue/river"
)

//go:embed migrations/*
var Migrations embed.FS

var (
	ErrInternalDB = errors.New(
		"an error occurred that was not possible to recover from",
	)
	ErrBeginTx             = errors.New("could not begin transaction")
	ErrRollbackTx          = errors.New("could not rollback transaction")
	ErrCommitTx            = errors.New("could not commit transaction")
	ErrNoRowWithIdentifier = errors.New(
		"could not find requested row in database",
	)
)

type Postgres struct {
	Pool  *pgxpool.Pool
	queue *river.Client[pgx.Tx]
}

func NewPostgres(dbPool *pgxpool.Pool, queue *river.Client[pgx.Tx]) Postgres {
	return Postgres{
		Pool:  dbPool,
		queue: queue,
	}
}

func (p *Postgres) Queue() *river.Client[pgx.Tx] {
	return p.queue
}

func (p *Postgres) NewQueue(opts ...queue.ClientCfgOpts) {
	p.queue = queue.NewClient(p.Pool, opts...)
}

func (p *Postgres) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "could not begin transaction", "reason", err)
		return nil, errors.Join(ErrBeginTx, err)
	}

	return tx, nil
}

func (p *Postgres) RollBackTx(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Rollback(ctx); err != nil {
		slog.ErrorContext(ctx, "could not rollback transaction", "reason", err)
		return errors.Join(ErrRollbackTx, err)
	}

	return nil
}

func (p *Postgres) CommitTx(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "could not commit transaction", "reason", err)
		return errors.Join(ErrCommitTx, err)
	}

	return nil
}

func CreatePooledConnection(
	ctx context.Context,
	uri string,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		slog.Error("could not parse database connection string", "error", err)
		return nil, err
	}

	// Add OpenTelemetry instrumentation
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("could not establish connection to database", "error", err)
		return nil, err
	}

	return dbpool, nil
}

func getFreePort() (uint32, error) {
	const (
		minPort = 1024
		maxPort = 65535
	)

	for range 10 {
		//nolint:gosec //only used for testing
		port := rand.Intn(maxPort-minPort) + minPort

		addr := fmt.Sprintf(":%d", port)
		conn, err := net.Listen("tcp", addr)
		if err != nil {
			continue // Port is in use, try another
		}

		conn.Close()
		//nolint:gosec //only used for testing
		return uint32(port), nil
	}

	return 0, fmt.Errorf("could not find an available port after 10 attempts")
}

type TestPostgres struct {
	Psql         Postgres
	EmbeddedPsql *embeddedpostgres.EmbeddedPostgres
	CleanupFunc  func()
}

func NewPostgresTest(
	ctx context.Context,
) (TestPostgres, error) {
	if config.Cfg.Environment == config.PROD_ENVIRONMENT {
		panic("don't NewPostgresTest in production")
	}

	user := "grafto"
	password := "grafto"
	database := fmt.Sprintf("grafto_test_%s", faker.DomainName())

	port, err := getFreePort()
	if err != nil {
		return TestPostgres{}, fmt.Errorf("failed to get free port: %w", err)
	}

	runtimePath := fmt.Sprintf("/tmp/psql_%s", uuid.New().String())
	runtimePathCleanup := func() {
		slog.Info("REMOVING EMBEDDED PSQL DIR")
		if err := os.RemoveAll(runtimePath); err != nil {
			slog.Error(
				"failed to remove temporary directory",
				"path",
				runtimePath,
				"error",
				err,
			)
		}
	}

	logger := &bytes.Buffer{}
	embeddedPsql := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Username(user).
			Password(password).
			Database(database).
			Version(embeddedpostgres.V16).
			Port(uint32(port)).
			RuntimePath(runtimePath).
			StartTimeout(45 * time.Second).
			StartParameters(map[string]string{"max_connections": "200"}).
			Logger(logger),
	)

	if err := embeddedPsql.Start(); err != nil {
		return TestPostgres{}, err
	}

	pool, err := CreatePooledConnection(
		ctx,
		fmt.Sprintf(
			"postgresql://%s:%s@localhost:%v/%s",
			user,
			password,
			port,
			database,
		),
	)
	if err != nil {
		return TestPostgres{}, err
	}

	db := stdlib.OpenDBFromPool(pool)

	gooseLock, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return TestPostgres{}, err
	}

	fsys, err := fs.Sub(Migrations, "migrations")
	if err != nil {
		return TestPostgres{}, err
	}
	gooseProvider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		fsys,
		goose.WithVerbose(false),
		goose.WithSessionLocker(gooseLock),
	)
	if err != nil {
		return TestPostgres{}, err
	}
	_, err = gooseProvider.Up(ctx)
	if err != nil {
		return TestPostgres{}, err
	}

	return TestPostgres{
		Psql: Postgres{
			Pool: pool,
		},
		EmbeddedPsql: embeddedPsql,
		CleanupFunc:  runtimePathCleanup,
	}, nil
}
