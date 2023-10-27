package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

func SetupDatabaseConnection(databaseURL string) *pgxpool.Pool {
	pgxConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		panic(err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		panic(err)
	}

	return dbpool
}

// ConvertTime converts a time.Time to a pgtype.Timestamptz until a fix is in place for sqlc
func ConvertTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
