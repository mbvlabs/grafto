package psql

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbv-labs/grafto/psql/database"
)

var (
	ErrInternalDBErr       = errors.New("an error occurred that was not possible to recover from")
	ErrNoRowWithIdentifier = errors.New("could not find requested row in database")
)

type Postgres struct {
	Queries *database.Queries
	tx      *pgxpool.Pool
}

func NewPostgres(dbPool *pgxpool.Pool) Postgres {
	return Postgres{
		Queries: database.New(dbPool),
		tx:      dbPool,
	}
}

func (p Postgres) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return p.tx.Begin(ctx)
}

func CreatePooledConnection(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, uri)
	if err != nil {
		slog.Error("could not establish connection to database", "error", err)
		return nil, err
	}

	return dbpool, nil
}
