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
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/mbvlabs/grafto/config"
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
	Queue *river.Client[pgx.Tx]
}

func NewPostgres(dbPool *pgxpool.Pool, queue *river.Client[pgx.Tx]) Postgres {
	return Postgres{
		dbPool,
		queue,
	}
}

func (p Postgres) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "could not begin transaction", "reason", err)
		return nil, errors.Join(ErrBeginTx, err)
	}

	return tx, nil
}

func (p Postgres) RollBackTx(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Rollback(ctx); err != nil {
		slog.ErrorContext(ctx, "could not rollback transaction", "reason", err)
		return errors.Join(ErrRollbackTx, err)
	}

	return nil
}

func (p Postgres) CommitTx(ctx context.Context, tx pgx.Tx) error {
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
	dbpool, err := pgxpool.New(ctx, uri)
	if err != nil {
		slog.Error("could not establish connection to database", "error", err)
		return nil, err
	}

	return dbpool, nil
}

// getFreePort returns a random available port number between 1024-65535
func getFreePort() (int, error) {
	const (
		minPort = 1024
		maxPort = 65535
	)

	for attempts := 0; attempts < 10; attempts++ {
		port := rand.Intn(maxPort-minPort) + minPort

		addr := fmt.Sprintf(":%d", port)
		conn, err := net.Listen("tcp", addr)
		if err != nil {
			continue // Port is in use, try another
		}

		conn.Close()
		return port, nil
	}

	return 0, fmt.Errorf("could not find an available port after 10 attempts")
}

func NewPostgresTest(
	ctx context.Context,
) (Postgres, *embeddedpostgres.EmbeddedPostgres, error) {
	if config.Cfg.Environment == config.PROD_ENVIRONMENT {
		panic("don't NewPostgresTest in production")
	}

	user := "grafto"
	password := "grafto"
	database := "grafto_test"

	port, err := getFreePort()
	if err != nil {
		return Postgres{}, nil, fmt.Errorf("failed to get free port: %w", err)
	}

	logger := &bytes.Buffer{}
	embeddedPsql := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Username(user).
			Password(password).
			Database(database).
			Version(embeddedpostgres.V16).
			Port(uint32(port)).
			RuntimePath("/tmp/psql").
			StartTimeout(45 * time.Second).
			StartParameters(map[string]string{"max_connections": "200"}).
			Logger(logger),
	)

	if err := embeddedPsql.Start(); err != nil {
		return Postgres{}, nil, err
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
		return Postgres{}, nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	gooseLock, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return Postgres{}, nil, err
	}

	fsys, err := fs.Sub(Migrations, "migrations")
	if err != nil {
		return Postgres{}, nil, err
	}
	gooseProvider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		fsys,
		goose.WithVerbose(true),
		goose.WithSessionLocker(gooseLock),
	)
	if err != nil {
		return Postgres{}, nil, err
	}
	_, err = gooseProvider.Up(ctx)
	if err != nil {
		return Postgres{}, nil, err
	}

	return Postgres{
		Pool: pool,
	}, embeddedPsql, nil
}
