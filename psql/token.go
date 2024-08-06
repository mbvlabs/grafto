package psql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbv-labs/grafto/psql/database"
)

func (p Postgres) DeleteTokenByHash(ctx context.Context, hash string) error {
	return p.Queries.DeleteTokenByHash(ctx, hash)
}

func (p Postgres) InsertToken(
	ctx context.Context,
	hash string,
	expiresAt time.Time,
	metaData []byte,
) error {
	return p.Queries.InsertToken(ctx, database.InsertTokenParams{
		ID: uuid.New(),
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Hash: hash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		MetaInformation: metaData,
	})
}

func (p Postgres) QueryTokenByHash(ctx context.Context, hash string) (database.Token, error) {
	return p.Queries.QueryTokenByHash(ctx, hash)
}
