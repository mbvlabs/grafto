package psql

import (
	"context"
	"embed"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
